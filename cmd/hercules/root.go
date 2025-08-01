package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"plugin"
	"runtime/pprof"
	"strings"
	"text/template"
	"unicode"

	"regexp"

	"github.com/Masterminds/sprig"
	progress "github.com/cheggaaa/pb/v3"
	"github.com/dmytrogajewski/hercules/api/proto/pb"
	"github.com/dmytrogajewski/hercules/internal/app/core"
	"github.com/dmytrogajewski/hercules/internal/pkg/config"
	"github.com/dmytrogajewski/hercules/internal/pkg/plumbing/uast"
	"github.com/dmytrogajewski/hercules/internal/pkg/version" // Using standard protobuf
	"github.com/go-git/go-billy/v6/osfs"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/cache"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/plumbing/transport/ssh"
	"github.com/go-git/go-git/v6/storage"
	"github.com/go-git/go-git/v6/storage/filesystem"
	"github.com/go-git/go-git/v6/storage/memory"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v2"
)

// oneLineWriter splits the output data by lines and outputs one on top of another using '\r'.
// It also does some dark magic to handle Git statuses.
type oneLineWriter struct {
	Writer io.Writer
}

func (writer oneLineWriter) Write(p []byte) (n int, err error) {
	strp := strings.TrimSpace(string(p))
	if strings.HasSuffix(strp, "done.") || len(strp) == 0 {
		strp = "cloning..."
	} else {
		strp = strings.Replace(strp, "\n", "\033[2K\r", -1)
	}
	_, err = writer.Writer.Write([]byte("\033[2K\r"))
	if err != nil {
		return
	}
	n, err = writer.Writer.Write([]byte(strp))
	return
}

func loadSSHIdentity(sshIdentity string) (*ssh.PublicKeys, error) {
	actual, err := homedir.Expand(sshIdentity)
	if err != nil {
		return nil, err
	}
	return ssh.NewPublicKeysFromFile("git", actual, "")
}

func loadRepository(uri string, cachePath string, disableStatus bool, sshIdentity string) *git.Repository {
	var repository *git.Repository
	var backend storage.Storer
	var err error

	if strings.Contains(uri, "://") || regexp.MustCompile("^[A-Za-z]\\w*@[A-Za-z0-9][\\w.]*:").MatchString(uri) {
		if cachePath != "" {
			backend = filesystem.NewStorage(osfs.New(cachePath), cache.NewObjectLRUDefault())
			_, err = os.Stat(cachePath)
			if !os.IsNotExist(err) {
				core.GetLogger().Warnf("warning: deleted %s\n", cachePath)
				os.RemoveAll(cachePath)
			}
		} else {
			backend = memory.NewStorage()
		}
		cloneOptions := &git.CloneOptions{URL: uri}
		if !disableStatus {
			fmt.Fprint(os.Stderr, "connecting...\r")
			cloneOptions.Progress = oneLineWriter{Writer: os.Stderr}
		}

		if sshIdentity != "" {
			auth, err := loadSSHIdentity(sshIdentity)
			if err != nil {
				core.GetLogger().Warnf("Failed loading SSH Identity %s\n", err)
			}
			cloneOptions.Auth = auth
		}

		repository, err = git.Clone(backend, nil, cloneOptions)
		if !disableStatus {
			fmt.Fprint(os.Stderr, "\033[2K\r")
		}
	} else if stat, err2 := os.Stat(uri); err2 == nil && !stat.IsDir() {
		// Temporarily disable siva filesystem support due to go-git v6 compatibility issues
		// if strings.HasSuffix(uri, ".siva") {
		// 	localFs := osfs.New(filepath.Dir(uri))
		// 	tmpFs := memfs.New()
		// 	basePath := filepath.Base(uri)
		// 	fs, err2 := sivafs.NewFilesystem(localFs, basePath, tmpFs)
		// 	if err2 != nil {
		// 		log.Panicf("unable to create a siva filesystem from %s: %v", uri, err2)
		// 	}
		// 	sivaStorage := filesystem.NewStorage(fs, cache.NewObjectLRUDefault())
		// 	repository, err = git.Open(sivaStorage, tmpFs)
		// } else {
		log.Panicf("siva filesystem support temporarily disabled due to go-git v6 compatibility issues")
		// }
	} else {
		if uri[len(uri)-1] == os.PathSeparator {
			uri = uri[:len(uri)-1]
		}
		repository, err = git.PlainOpen(uri)
	}
	if err != nil {
		log.Panicf("failed to open %s: %v", uri, err)
	}
	return repository
}

type arrayPluginFlags map[string]bool

func (apf *arrayPluginFlags) String() string {
	var list []string
	for key := range *apf {
		list = append(list, key)
	}
	return strings.Join(list, ", ")
}

func (apf *arrayPluginFlags) Set(value string) error {
	(*apf)[value] = true
	return nil
}

func (apf *arrayPluginFlags) Type() string {
	return "path"
}

func loadPlugins() {
	pluginFlags := arrayPluginFlags{}
	fs := pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)
	fs.SetOutput(ioutil.Discard)
	pluginFlagName := "plugin"
	const pluginDesc = "Load the specified plugin by the full or relative path. " +
		"Can be specified multiple times."
	fs.Var(&pluginFlags, pluginFlagName, pluginDesc)
	err := cobra.MarkFlagFilename(fs, "plugin")
	if err != nil {
		panic(err)
	}
	pflag.Var(&pluginFlags, pluginFlagName, pluginDesc)
	fs.Parse(os.Args[1:])
	for path := range pluginFlags {
		_, err := plugin.Open(path)
		if err != nil {
			core.GetLogger().Warnf("Failed to load plugin from %s %s\n", path, err)
		}
	}
}

// Helper: returns true if all dependencies are satisfied by the provided set
func dependenciesSatisfied(requires []string, provided map[string]struct{}) bool {
	for _, dep := range requires {
		if _, ok := provided[dep]; !ok {
			return false
		}
	}
	return true
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hercules",
	Short: "Analyse a Git repository.",
	Long: `Hercules is a flexible and fast Git repository analysis engine. The base command executes
the commit processing pipeline which is automatically generated from the dependencies of one
or several analysis targets. The list of the available targets is printed in --help. External
targets can be added using the --plugin system.`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		leaves := core.Registry.GetLeaves()
		flags := cmd.Flags()
		getBool := func(name string) bool {
			value, err := flags.GetBool(name)
			if err != nil {
				panic(err)
			}
			return value
		}
		getString := func(name string) string {
			value, err := flags.GetString(name)
			if err != nil {
				panic(err)
			}
			return value
		}
		firstParent := getBool("first-parent")
		commitsFile := getString("commits")
		head := getBool("head")
		protobuf := getBool("pb")
		json := getBool("json")
		yamlFlag := getBool("yaml")
		profile := getBool("profile")
		disableStatus := getBool("quiet")
		sshIdentity := getString("ssh-identity")
		allAnalyses := getBool("all")
		uastProvider := getString("uast-provider")

		if profile {
			go func() {
				err := http.ListenAndServe("localhost:6060", nil)
				if err != nil {
					panic(err)
				}
			}()
			prof, _ := os.Create("hercules.pprof")
			err := pprof.StartCPUProfile(prof)
			if err != nil {
				panic(err)
			}
			defer pprof.StopCPUProfile()
		}
		uri := args[0]
		cachePath := ""
		if len(args) == 2 {
			cachePath = args[1]
		}
		repository := loadRepository(uri, cachePath, disableStatus, sshIdentity)

		// core logic
		if cmdlineDeployed == nil {
			core.GetLogger().Infof("[DEBUG] cmdlineDeployed was nil, initializing empty map.")
			cmdlineDeployed = make(map[string]*bool)
		}
		pipeline := core.NewPipeline(repository)
		pipeline.SetFeaturesFromFlags()
		var bar *progress.ProgressBar
		if !disableStatus {
			pipeline.OnProgress = func(commit, length int, action string) {
				if bar == nil {
					bar = progress.New(length)
					// Progress bar API changed - simplified
					bar.Start()
				}
				if action == core.MessageFinalize {
					bar.Finish()
					fmt.Fprint(os.Stderr, "\033[2K\rfinalizing...")
				}
			}
		}

		var commits []*object.Commit
		var err error
		if commitsFile == "" {
			if !head {
				fmt.Fprint(os.Stderr, "git log...\r")
				commits, err = pipeline.Commits(firstParent)
			} else {
				commits, err = pipeline.HeadCommit()
			}
		} else {
			commits, err = core.LoadCommitsFromFile(commitsFile, repository)
		}
		if err != nil {
			log.Fatalf("failed to list the commits: %v", err)
		}
		cmdlineFacts[core.ConfigPipelineCommits] = commits
		if uastProvider != "" {
			cmdlineFacts[uast.ConfigUASTProvider] = uastProvider
		}
		if allAnalyses {
			repository := loadRepository(uri, cachePath, disableStatus, sshIdentity)
			// Deploy all leaves and all plumbing items in a single pipeline
			core.Registry = core.NewPipelineItemRegistry()
			cmdlineDeployed := map[string]*bool{}
			b := true
			for _, leaf := range leaves {
				cmdlineDeployed[leaf.Name()] = &b
			}
			pipeline := core.NewPipeline(repository)
			pipeline.SetFeaturesFromFlags()
			// Deploy all plumbing items
			for _, item := range core.Registry.GetPlumbingItems() {
				pipeline.DeployItem(item)
			}
			// Deploy all leaves
			for _, leaf := range leaves {
				pipeline.DeployItem(leaf)
			}
			err := pipeline.Initialize(cmdlineFacts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] Failed to initialize pipeline for --all: %v\n", err)
				return
			}
			results, err := pipeline.Run(commits)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] Failed to run pipeline for --all: %v\n", err)
				return
			}
			for _, leaf := range leaves {
				result := results[leaf]
				fmt.Printf("%s:\n", leaf.Name())
				if err := leaf.Serialize(result, false, os.Stdout); err != nil {
					panic(err)
				}
			}
			return
		}

		var deployed []core.LeafPipelineItem
		// Handle individual analysis flags
		for _, leaf := range leaves {
			flagName := leaf.Flag()
			if getBool(flagName) {
				b := true
				cmdlineDeployed[leaf.Name()] = &b
				pipeline.DeployItem(leaf)
			}
		}
		// Debug: Print pipeline items and their dependencies
		// fmt.Println("[DEBUG] Pipeline items to be initialized:")
		// for _, item := range pipeline.Items() {
		// 	fmt.Printf("  - %s\n", item.Name())
		// 	fmt.Printf("    Provides: %v\n", item.Provides())
		// 	fmt.Printf("    Requires: %v\n", item.Requires())
		// }
		err = pipeline.Initialize(cmdlineFacts)
		if err != nil {
			log.Fatal(err)
		}
		// After initialization, collect deployed leaves
		deployed = nil
		for _, item := range pipeline.Items() {
			if leaf, ok := item.(core.LeafPipelineItem); ok {
				deployed = append(deployed, leaf)
			}
		}
		results, err := pipeline.Run(commits)
		if err != nil {
			log.Fatalf("failed to run the pipeline: %v", err)
		}
		if !disableStatus {
			fmt.Fprint(os.Stderr, "\033[2K\r")
			// if not a terminal, the user will not see the output, so show the status
			if !term.IsTerminal(int(os.Stdout.Fd())) {
				fmt.Fprint(os.Stderr, "writing...\r")
			}
		}
		// Output format precedence: pb > json > yaml (default)
		if protobuf && (json || yamlFlag) {
			log.Fatal("--pb cannot be combined with --json or --yaml")
		}
		if json && yamlFlag {
			log.Fatal("--json and --yaml cannot be used together")
		}
		if protobuf {
			protobufResults(uri, deployed, results)
		} else if json {
			viper.Set("hercules.json_output", true)
			jsonResults(uri, deployed, results)
		} else {
			printResults(uri, deployed, results)
		}
	},
}

func printResults(
	uri string, deployed []core.LeafPipelineItem,
	results map[core.LeafPipelineItem]interface{}) {
	commonResult := results[nil].(*core.CommonAnalysisResult)

	fmt.Println("hercules:")
	fmt.Printf("  version: %d\n", version.Binary)
	fmt.Println("  hash:", version.BinaryGitHash)
	fmt.Println("  repository:", uri)
	fmt.Println("  begin_unix_time:", commonResult.BeginTime)
	fmt.Println("  end_unix_time:", commonResult.EndTime)
	fmt.Println("  commits:", commonResult.CommitsNumber)
	fmt.Println("  run_time:", commonResult.RunTime.Nanoseconds()/1e6)

	// fmt.Fprintln(os.Stderr, "[DEBUG] Results map keys:")
	// for k := range results {
	// 	if k == nil {
	// 		fmt.Fprintln(os.Stderr, "  <nil>")
	// 	} else {
	// 		fmt.Fprintln(os.Stderr, "  ", k.Name())
	// 	}
	// }
	// fmt.Fprintln(os.Stderr, "[DEBUG] Deployed leaf names:")
	// for _, item := range deployed {
	// 	fmt.Fprintln(os.Stderr, "  ", item.Name())
	// }
	// os.Stderr.Sync()
	// fmt.Fprintf(os.Stderr, "[DEBUG] Results map len: %d\n", len(results))
	// fmt.Fprintf(os.Stderr, "[DEBUG] Deployed slice len: %d\n", len(deployed))
	if len(results) == 0 {
		// fmt.Fprintln(os.Stderr, "[DEBUG] Results map is empty!")
		panic("Results map is empty after pipeline.Run; check pipeline deployment and run logic.")
	}
	// Build a map from leaf name to result
	leafResults := make(map[string]interface{})
	for k, v := range results {
		if k == nil {
			continue
		}
		leafResults[k.Name()] = v
	}
	// fmt.Fprintln(os.Stderr, "[DEBUG] leafResults map keys:")
	// for k := range leafResults {
	// 	fmt.Fprintln(os.Stderr, "  ", k)
	// }
	for _, item := range deployed {
		result, ok := leafResults[item.Name()]
		if !ok {
			fmt.Fprintf(os.Stderr, "[WARNING] No result for analysis %s\n", item.Name())
			continue
		}
		fmt.Printf("%s:\n", item.Name())
		if err := item.Serialize(result, false, os.Stdout); err != nil {
			panic(err)
		}
	}
}

func protobufResults(
	uri string, deployed []core.LeafPipelineItem,
	results map[core.LeafPipelineItem]interface{}) {

	header := pb.Metadata{
		Version:    2,
		Hash:       version.BinaryGitHash,
		Repository: uri,
	}
	results[nil].(*core.CommonAnalysisResult).FillMetadata(&header)

	message := pb.AnalysisResults{
		Header:   &header,
		Contents: map[string][]byte{},
	}

	for _, item := range deployed {
		result := results[item]
		buffer := &bytes.Buffer{}
		if err := item.Serialize(result, true, buffer); err != nil {
			panic(err)
		}
		message.Contents[item.Name()] = buffer.Bytes()
	}

	serialized, err := proto.Marshal(&message)
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(serialized)
}

func yamlToJSONCompatible(v interface{}) interface{} {
	// Recursively convert map[interface{}]interface{} to map[string]interface{}
	switch x := v.(type) {
	case map[interface{}]interface{}:
		m2 := make(map[string]interface{}, len(x))
		for k, v2 := range x {
			m2[fmt.Sprint(k)] = yamlToJSONCompatible(v2)
		}
		return m2
	case []interface{}:
		for i, v2 := range x {
			x[i] = yamlToJSONCompatible(v2)
		}
		return x
	default:
		return x
	}
}

func jsonResults(
	uri string, deployed []core.LeafPipelineItem,
	results map[core.LeafPipelineItem]interface{}) {

	commonResult := results[nil].(*core.CommonAnalysisResult)

	// Create the main JSON structure
	output := map[string]interface{}{
		"hercules": map[string]interface{}{
			"version":         version.Binary,
			"hash":            version.BinaryGitHash,
			"repository":      uri,
			"begin_unix_time": commonResult.BeginTime,
			"end_unix_time":   commonResult.EndTime,
			"commits":         commonResult.CommitsNumber,
			"run_time":        commonResult.RunTime.Nanoseconds() / 1e6,
		},
	}

	// Add results for each deployed item
	for _, item := range deployed {
		result := results[item]
		buffer := &bytes.Buffer{}
		if err := item.Serialize(result, false, buffer); err != nil {
			panic(err)
		}
		var yamlData interface{}
		if err := yaml.Unmarshal(buffer.Bytes(), &yamlData); err != nil {
			output[item.Name()] = buffer.String()
		} else {
			output[item.Name()] = yamlToJSONCompatible(yamlData)
		}
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(jsonData)
}

// trimRightSpace removes the trailing whitespace characters.
func trimRightSpace(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

// rpad adds padding to the right of a string.
func rpad(s string, padding int) string {
	return fmt.Sprintf(fmt.Sprintf("%%-%ds", padding), s)
}

// tmpl was adapted from cobra/cobra.go
func tmpl(w io.Writer, text string, data interface{}) error {
	var templateFuncs = template.FuncMap{
		"trim":                    strings.TrimSpace,
		"trimRightSpace":          trimRightSpace,
		"trimTrailingWhitespaces": trimRightSpace,
		"rpad":                    rpad,
		"gt":                      cobra.Gt,
		"eq":                      cobra.Eq,
	}
	for k, v := range sprig.TxtFuncMap() {
		templateFuncs[k] = v
	}
	t := template.New("top")
	t.Funcs(templateFuncs)
	template.Must(t.Parse(text))
	return t.Execute(w, data)
}

func formatUsage(c *cobra.Command) error {
	// the default UsageFunc() does some private magic c.mergePersistentFlags()
	// this should stay on top
	localFlags := c.LocalFlags()
	leaves := core.Registry.GetLeaves()
	plumbing := core.Registry.GetPlumbingItems()
	features := core.Registry.GetFeaturedItems()
	core.EnablePathFlagTypeMasquerade()
	filter := map[string]bool{}
	for _, l := range leaves {
		filter[l.Flag()] = true
		for _, cfg := range l.ListConfigurationOptions() {
			filter[cfg.Flag] = true
		}
	}
	for _, i := range plumbing {
		for _, cfg := range i.ListConfigurationOptions() {
			filter[cfg.Flag] = true
		}
	}

	for key := range filter {
		localFlags.Lookup(key).Hidden = true
	}
	args := map[string]interface{}{
		"c":        c,
		"leaves":   leaves,
		"plumbing": plumbing,
		"features": features,
	}

	helpTemplate := `Usage:{{if .c.Runnable}}
  {{.c.UseLine}}{{end}}{{if .c.HasAvailableSubCommands}}
  {{.c.CommandPath}} [command]{{end}}{{if gt (len .c.Aliases) 0}}

Aliases:
  {{.c.NameAndAliases}}{{end}}{{if .c.HasExample}}

Examples:
{{.c.Example}}{{end}}{{if .c.HasAvailableSubCommands}}

Available Commands:{{range .c.Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .c.HasAvailableLocalFlags}}

Flags:
{{range $line := .c.LocalFlags.FlagUsages | trimTrailingWhitespaces | split "\n"}}
{{- $desc := splitList "   " $line | last}}
{{- $offset := sub ($desc | len) ($desc | trim | len)}}
{{- $indent := splitList "   " $line | initial | join "   " | len | add 3 | add $offset | int}}
{{- $wrap := sub 120 $indent | int}}
{{- splitList "   " $line | initial | join "   "}}   {{cat "!" $desc | wrap $wrap | indent $indent | substr $indent -1 | substr 2 -1}}
{{end}}{{end}}

Analysis Targets:{{range .leaves}}
      --{{rpad .Flag 40}}Runs {{.Name}} analysis.{{wrap 72 .Description | nindent 48}}{{range .ListConfigurationOptions}}
          --{{if .Type.String}}{{rpad (print .Flag " " .Type.String) 40}}{{else}}{{rpad .Flag 40}}{{end}}
          {{- $desc := dict "desc" .Description}}
          {{- if .Default}}{{$_ := set $desc "desc" (print .Description " The default value is " .FormatDefault ".")}}
          {{- end}}
          {{- $desc := pluck "desc" $desc | first}}
          {{- $desc | wrap 68 | indent 52 | substr 52 -1}}{{end}}
{{end}}

Plumbing Options:{{range .plumbing}}{{$name := .Name}}{{range .ListConfigurationOptions}}
      --{{if .Type.String}}{{rpad (print .Flag " " .Type.String " [" $name "]") 40}}{{else}}{{rpad (print .Flag " [" $name "]") 40}}
        {{- end}}
        {{- $desc := dict "desc" .Description}}
        {{- if .Default}}{{$_ := set $desc "desc" (print .Description " The default value is " .FormatDefault ".")}}
        {{- end}}
        {{- $desc := pluck "desc" $desc | first}}{{$desc | wrap 72 | indent 48 | substr 48 -1}}{{end}}{{end}}

--feature:{{range $key, $value := .features}}
      {{rpad $key 42}}Enables {{range $index, $item := $value}}{{if $index}}, {{end}}{{$item.Name}}{{end}}.{{end}}{{if .c.HasAvailableInheritedFlags}}

Global Flags:
{{.c.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .c.HasHelpSubCommands}}

Additional help topics:{{range .c.Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .c.HasAvailableSubCommands}}

Use "{{.c.CommandPath}} [command] --help" for more information about a command.{{end}}
`
	err := tmpl(c.OutOrStderr(), helpTemplate, args)
	for key := range filter {
		localFlags.Lookup(key).Hidden = false
	}
	if err != nil {
		c.Println(err)
	}
	return err
}

// versionCmd prints the API version and the Git commit hash
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information and exit.",
	Long:  ``,
	Args:  cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %d\nGit:     %s\n", version.Binary, version.BinaryGitHash)
	},
}

var cmdlineFacts map[string]interface{}
var cmdlineDeployed map[string]*bool

func init() {
	loadPlugins()
	rootFlags := rootCmd.Flags()
	rootFlags.String("commits", "", "Path to the text file with the "+
		"commit history to follow instead of the default 'git log'. "+
		"The format is the list of hashes, each hash on a "+
		"separate line. The first hash is the root.")
	err := rootCmd.MarkFlagFilename("commits")
	if err != nil {
		panic(err)
	}
	core.PathifyFlagValue(rootFlags.Lookup("commits"))
	rootFlags.Bool("head", false, "Analyze only the latest commit.")
	rootFlags.Bool("first-parent", false, "Follow only the first parent in the commit history - "+
		"\"git log --first-parent\".")
	rootFlags.Bool("pb", false, "The output format will be Protocol Buffers instead of YAML.")
	rootFlags.Bool("json", false, "The output format will be JSON instead of YAML.")
	rootFlags.Bool("yaml", false, "The output format will be YAML (default, mutually exclusive with --json and --pb).")
	rootFlags.Bool("quiet", !term.IsTerminal(int(os.Stdin.Fd())),
		"Do not print status updates to stderr.")
	rootFlags.Bool("profile", false, "Collect the profile to hercules.pprof.")
	rootFlags.Bool("all", false, "Run all available analyses (mutually exclusive with individual analysis flags).")
	rootFlags.String("uast-provider", "", "UAST provider to use (embedded, babelfish, etc.)")
	rootFlags.String("ssh-identity", "", "Path to SSH identity file (e.g., ~/.ssh/id_rsa) to clone from an SSH remote.")
	err = rootCmd.MarkFlagFilename("ssh-identity")
	if err != nil {
		panic(err)
	}
	core.PathifyFlagValue(rootFlags.Lookup("ssh-identity"))
	rootFlags.String("log-file", "", "Path to log file. If not set, logging is disabled.")
	rootFlags.String("log-format", "plain", "Log format: 'plain' or 'json'. Default is 'plain'.")
	cmdlineFacts, cmdlineDeployed = core.Registry.AddFlags(rootFlags)
	rootCmd.SetUsageFunc(formatUsage)
	rootCmd.AddCommand(versionCmd)
	versionCmd.SetUsageFunc(versionCmd.UsageFunc())
}

func selectLogger(serverMode bool, logFile string) core.Logger {
	if serverMode {
		return core.NewSlogLogger(os.Stdout)
	} else if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			return core.NewFileLogger(f)
		} else {
			return &core.NoOpLogger{}
		}
	} else {
		return &core.NoOpLogger{}
	}
}

func main() {
	// Load config
	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to load config:", err)
		os.Exit(1)
	}

	// Detect server mode
	serverMode := cfg.GRPC.Enabled || cfg.Server.Enabled

	// Parse --log-file flag
	logFile := ""
	for i, arg := range os.Args {
		if arg == "--log-file" && i+1 < len(os.Args) {
			logFile = os.Args[i+1]
		}
	}

	// Debug output
	fmt.Fprintf(os.Stderr, "DEBUG: serverMode=%v, logFile=%q\n", serverMode, logFile)

	logger := selectLogger(serverMode, logFile)
	core.SetLogger(logger)

	// Debug output
	fmt.Fprintf(os.Stderr, "DEBUG: selected logger type: %T\n", logger)

	if err := rootCmd.Execute(); err != nil {
		logger.Error(os.Stderr, err)
		os.Exit(1)
	}
}
