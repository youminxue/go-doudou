package cmd

import (
	"github.com/spf13/cobra"
	"github.com/youminxue/v2/cmd/internal/name"
)

var file string
var strategy string
var omitempty bool
var form bool

// nameCmd updates json tag of struct fields
var nameCmd = &cobra.Command{
	Use:   "name",
	Short: "bulk add or update json tag of struct fields",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		n := name.Name{file, strategy, omitempty, form}
		n.Exec()
	},
}

func init() {
	rootCmd.AddCommand(nameCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// nameCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	nameCmd.Flags().StringVarP(&file, "file", "f", "", "absolute path of dto file")
	nameCmd.Flags().StringVarP(&strategy, "strategy", "s", "lowerCamel", `name of strategy, currently only support "lowerCamel" and "snake"`)
	nameCmd.Flags().BoolVarP(&omitempty, "omitempty", "o", false, "whether omit empty value or not")
	nameCmd.Flags().BoolVar(&form, "form", false, "whether need form tag for https://github.com/go-playground/form")
}
