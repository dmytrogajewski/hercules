package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"sort"
	"strings"

	progress "github.com/cheggaaa/pb/v3"
	"github.com/dmytrogajewski/hercules/api/proto/pb"
	"github.com/dmytrogajewski/hercules/internal/app/core"
	"github.com/dmytrogajewski/hercules/internal/pkg/version"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

// combineCmd represents the combine command
var combineCmd = &cobra.Command{
	Use:   "combine",
	Short: "Merge several binary analysis results together.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, files []string) {
		if len(files) == 1 {
			file, err := os.Open(files[0])
			if err != nil {
				panic(err)
			}
			defer file.Close()
			_, err = io.Copy(os.Stdout, bufio.NewReader(file))
			if err != nil {
				panic(err)
			}
			return
		}
		only, err := cmd.Flags().GetString("only")
		if err != nil {
			panic(err)
		}
		var repos []string
		allErrors := map[string][]string{}
		mergedResults := map[string]interface{}{}
		mergedMetadata := &core.CommonAnalysisResult{}
		var fileName string
		bar := progress.New(len(files))
		// Progress bar callback removed - API changed
		bar.Start()
		debug.SetGCPercent(20)
		for _, fileName = range files {
			bar.Increment()
			anotherResults, anotherMetadata, errs := loadMessage(fileName, &repos)
			if anotherMetadata != nil {
				mergeErrs := mergeResults(mergedResults, mergedMetadata, anotherResults, anotherMetadata, only)
				for _, err := range mergeErrs {
					errs = append(errs, err.Error())
				}
			}
			allErrors[fileName] = errs
			debug.FreeOSMemory()
		}
		bar.Finish()
		os.Stderr.WriteString("\033[2K\r")
		printErrors(allErrors)
		sort.Strings(repos)
		if mergedMetadata == nil {
			return
		}

		mergedMessage := pb.AnalysisResults{
			Header: &pb.Metadata{
				Version:    int32(version.Binary),
				Hash:       version.BinaryGitHash,
				Repository: strings.Join(repos, " & "),
			},
			Contents: map[string][]byte{},
		}
		mergedMetadata.FillMetadata(mergedMessage.Header)
		for key, val := range mergedResults {
			buffer := bytes.Buffer{}
			err := core.Registry.Summon(key)[0].(core.LeafPipelineItem).Serialize(
				val, true, &buffer)
			if err != nil {
				panic(err)
			}
			mergedMessage.Contents[key] = buffer.Bytes()
		}
		serialized, err := proto.Marshal(&mergedMessage)
		if err != nil {
			panic(err)
		}
		os.Stdout.Write(serialized)
	},
}

func loadMessage(fileName string, repos *[]string) (
	map[string]interface{}, *core.CommonAnalysisResult, []string) {

	var errs []string
	fi, err := os.Stat(fileName)
	if err != nil {
		errs = append(errs, "Cannot access "+fileName+": "+err.Error())
		return nil, nil, errs
	}
	if fi.Size() == 0 {
		errs = append(errs, "Cannot parse "+fileName+": file size is 0")
		return nil, nil, errs
	}
	buffer, err := os.ReadFile(fileName)
	if err != nil {
		errs = append(errs, "Cannot read "+fileName+": "+err.Error())
		return nil, nil, errs
	}
	message := pb.AnalysisResults{}
	err = proto.Unmarshal(buffer, &message)
	if err != nil {
		errs = append(errs, "Cannot parse "+fileName+": "+err.Error())
		return nil, nil, errs
	}
	if message.Header == nil {
		errs = append(errs, "Cannot parse "+fileName+": corrupted header")
		return nil, nil, errs
	}
	*repos = append(*repos, message.Header.Repository)
	results := map[string]interface{}{}
	for key, val := range message.Contents {
		summoned := core.Registry.Summon(key)
		if len(summoned) == 0 {
			errs = append(errs, fileName+": item not found: "+key)
			continue
		}
		mpi, ok := summoned[0].(core.ResultMergeablePipelineItem)
		if !ok {
			errs = append(errs, fileName+": "+key+": ResultMergeablePipelineItem is not implemented")
			continue
		}
		msg, err := mpi.Deserialize(val)
		if err != nil {
			errs = append(errs, fileName+": deserialization failed: "+key+": "+err.Error())
			continue
		}
		results[key] = msg
	}
	return results, core.MetadataToCommonAnalysisResult(message.Header), errs
}

func printErrors(allErrors map[string][]string) {
	needToPrintErrors := false
	for _, errs := range allErrors {
		if len(errs) > 0 {
			needToPrintErrors = true
			break
		}
	}
	if !needToPrintErrors {
		return
	}
	fmt.Fprintln(os.Stderr, "Errors:")
	for key, errs := range allErrors {
		if len(errs) > 0 {
			fmt.Fprintln(os.Stderr, "  "+key)
			for _, err := range errs {
				fmt.Fprintln(os.Stderr, "    "+err)
			}
		}
	}
}

func mergeResults(mergedResults map[string]interface{},
	mergedCommons *core.CommonAnalysisResult,
	anotherResults map[string]interface{},
	anotherCommons *core.CommonAnalysisResult,
	only string) []error {
	var errors []error
	for key, val := range anotherResults {
		if only != "" && key != only {
			continue
		}
		mergedResult, exists := mergedResults[key]
		if !exists {
			mergedResults[key] = val
			continue
		}
		item := core.Registry.Summon(key)[0].(core.ResultMergeablePipelineItem)
		mergedResult = item.MergeResults(mergedResult, val, mergedCommons, anotherCommons)
		if err, isErr := mergedResult.(error); isErr {
			errors = append(errors, fmt.Errorf("could not merge %s: %v", item.Name(), err))
		} else {
			mergedResults[key] = mergedResult
		}
	}
	if mergedCommons.CommitsNumber == 0 {
		*mergedCommons = *anotherCommons
	} else {
		mergedCommons.Merge(anotherCommons)
	}
	return errors
}

func getOptionsString() string {
	var leaves []string
	for _, leaf := range core.Registry.GetLeaves() {
		leaves = append(leaves, leaf.Name())
	}
	return strings.Join(leaves, ", ")
}

func init() {
	rootCmd.AddCommand(combineCmd)
	combineCmd.SetUsageFunc(combineCmd.UsageFunc())
	combineCmd.Flags().String("only", "", "Consider only the specified analysis. "+
		"Empty means all available. Choices: "+getOptionsString()+".")
}
