package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	command "github.com/beringresearch/macpine/cmd"

	"github.com/spf13/cobra"
)

const descriptionSourcePath = "docs/docs/cli/"

func printOptions(buf *bytes.Buffer, cmd *cobra.Command, name string) error {
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		buf.WriteString("## Options\n\n```\n")
		flags.PrintDefaults()
		buf.WriteString("```\n\n")
	}

	parentFlags := cmd.InheritedFlags()
	parentFlags.SetOutput(buf)
	if parentFlags.HasAvailableFlags() {
		buf.WriteString("## Options inherited from parent commands\n\n```\n")
		parentFlags.PrintDefaults()
		buf.WriteString("```\n\n")
	}
	return nil
}

// GenMarkdown creates markdown output.
func GenMarkdown(cmd *cobra.Command, w io.Writer) error {
	return GenMarkdownCustom(cmd, w, func(s string) string { return s })
}

// GenMarkdownCustom creates custom markdown output.
func GenMarkdownCustom(cmd *cobra.Command, w io.Writer, linkHandler func(string) string) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)
	name := cmd.CommandPath()

	short := cmd.Short
	long := cmd.Long
	if len(long) == 0 {
		long = short
	}

	buf.WriteString("# " + name + "\n\n")
	buf.WriteString(short + "\n\n")
	if cmd.Runnable() {
		buf.WriteString(fmt.Sprintf("```\n%s\n```\n\n", cmd.UseLine()))
	}

	buf.WriteString("## Description\n\n")

	buf.WriteString(long + "\n\n")

	if len(cmd.Example) > 0 {
		buf.WriteString("## Examples\n\n")
		buf.WriteString(cmd.Example + "\n\n")
	}

	if err := printOptions(buf, cmd, name); err != nil {
		return err
	}

	//buf.WriteString("## See Also\n\n")
	//if cmd.HasParent() {
	//	parent := cmd.Parent()
	//	pname := parent.CommandPath()
	//	link := pname + ".md"
	//	link = strings.Replace(link, " ", "_", -1)
	//	buf.WriteString(fmt.Sprintf("* [%s](%s)\t - %s\n", pname, linkHandler(link), parent.Short))
	//	cmd.VisitParents(func(c *cobra.Command) {
	//		if c.DisableAutoGenTag {
	//			cmd.DisableAutoGenTag = c.DisableAutoGenTag
	//		}
	//	})
	//}

	//children := cmd.Commands()
	//sort.Sort(byName(children))

	//for _, child := range children {
	//	if !child.IsAvailableCommand() || child.IsAdditionalHelpTopicCommand() {
	//		continue
	//	}
	//	cname := name + " " + child.Name()
	//	link := cname + ".md"
	//	link = strings.Replace(link, " ", "_", -1)
	//	buf.WriteString(fmt.Sprintf("* [%s](%s)\t - %s\n", cname, linkHandler(link), child.Short))
	//}
	//buf.WriteString("\n")

	_, err := buf.WriteTo(w)
	return err
}

// GenMarkdownTree will generate a markdown page for this command and all
// descendants in the directory given. The header may be nil.
// This function may not work correctly if your command names have `-` in them.
// If you have `cmd` with two subcmds, `sub` and `sub-third`,
// and `sub` has a subcommand called `third`, it is undefined which
// help output will be in the file `cmd-sub-third.1`.
func GenMarkdownTree(cmd *cobra.Command, dir string) error {
	identity := func(s string) string { return s }
	emptyStr := func(s string) string { return "" }
	return GenMarkdownTreeCustom(cmd, dir, emptyStr, identity)
}

// GenMarkdownTreeCustom is the the same as GenMarkdownTree, but
// with custom filePrepender and linkHandler.
func GenMarkdownTreeCustom(cmd *cobra.Command, dir string, filePrepender, linkHandler func(string) string) error {
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := GenMarkdownTreeCustom(c, dir, filePrepender, linkHandler); err != nil {
			return err
		}
	}

	basename := strings.Replace(cmd.CommandPath(), " ", "_", -1) + ".md"
	filename := filepath.Join(dir, basename)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.WriteString(f, filePrepender(filename)); err != nil {
		return err
	}
	if err := GenMarkdownCustom(cmd, f, linkHandler); err != nil {
		return err
	}
	return nil
}

func main() {
	rootCmd := command.MacpineCmd
	disableFlagsInUseLine(rootCmd)
	source := filepath.Join(descriptionSourcePath)
	fmt.Println("Markdown source:", source)

	if err := loadLongDescription(rootCmd, source); err != nil {
		log.Fatal(err)
	}

	rootCmd.DisableAutoGenTag = true
	err := GenMarkdownTree(rootCmd, "./docs/cli")
	if err != nil {
		log.Fatal(err)
	}
}

func disableFlagsInUseLine(cmd *cobra.Command) {
	visitAll(cmd, func(ccmd *cobra.Command) {
		// do not add a `[flags]` to the end of the usage line.
		ccmd.DisableFlagsInUseLine = true
	})
}

// visitAll will traverse all commands from the root.
// This is different from the VisitAll of cobra.Command where only parents
// are checked.
func visitAll(root *cobra.Command, fn func(*cobra.Command)) {
	for _, cmd := range root.Commands() {
		visitAll(cmd, fn)
	}
	fn(root)
}

func loadLongDescription(parentCmd *cobra.Command, path string) error {
	for _, cmd := range parentCmd.Commands() {
		if cmd.HasSubCommands() {
			if err := loadLongDescription(cmd, path); err != nil {
				return err
			}
		}
		name := cmd.CommandPath()
		log.Println("INFO: Generating docs for", name)
		if i := strings.Index(name, " "); i >= 0 {
			// remove root command / binary name
			name = name[i+1:]
		}
		if name == "" {
			continue
		}
		mdFile := "brave_" + strings.ReplaceAll(name, " ", "_") + ".md"

		fullPath := filepath.Join(path, mdFile)
		content, err := ioutil.ReadFile(fullPath)
		if os.IsNotExist(err) {
			log.Printf("WARN: %s does not exist, skipping\n", mdFile)
			continue
		}
		if err != nil {
			return err
		}
		description, examples := parseMDContent(string(content))
		cmd.Long = description
		cmd.Example = examples
	}
	return nil
}

func parseMDContent(mdString string) (description string, examples string) {
	parsedContent := strings.Split(mdString, "\n## ")
	for _, s := range parsedContent {
		if strings.Index(s, "Description") == 0 {
			description = strings.TrimSpace(strings.TrimPrefix(s, "Description"))
		}
		if strings.Index(s, "Examples") == 0 {
			examples = strings.TrimSpace(strings.TrimPrefix(s, "Examples"))
		}
	}
	return description, examples
}

type byName []*cobra.Command

func (s byName) Len() int           { return len(s) }
func (s byName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byName) Less(i, j int) bool { return s[i].Name() < s[j].Name() }
