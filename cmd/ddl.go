package cmd

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/youminxue/odin/cmd/internal/ddl"
	"github.com/youminxue/odin/cmd/internal/ddl/config"
	"github.com/youminxue/odin/toolkit/dotenv"
	"github.com/youminxue/odin/toolkit/pathutils"
	"github.com/youminxue/odin/toolkit/yaml"
)

var dir string
var reverse bool
var dao bool
var pre string
var df string
var env string

// ddlCmd generates entity and dao layer source code from database tables and update tables from entity code
var ddlCmd = &cobra.Command{
	Use:   "ddl",
	Short: "migration tool between database table structure and golang struct",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		yaml.Load(env)
		dotenv.Load(env)
		var conf config.DbConfig
		err := envconfig.Process("db", &conf)
		if err != nil {
			logrus.Panicln("Error processing env", err)
		}
		if dir, err = pathutils.FixPath(dir, "entity"); err != nil {
			logrus.Panicln(err)
		}
		d := ddl.Ddl{dir, reverse, dao, pre, df, conf}
		d.Exec()
	},
}

func init() {
	rootCmd.AddCommand(ddlCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ddlCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	ddlCmd.Flags().StringVar(&dir, "entity", "entity", "Path of entity folder.")
	ddlCmd.Flags().StringVar(&pre, "pre", "", "Table name prefix. e.g.: prefix biz_ for biz_product.")
	ddlCmd.Flags().StringVar(&df, "df", "dao", "Name of dao folder.")
	ddlCmd.Flags().StringVar(&env, "env", "dev", "Environment name such as dev, uat, test, prod, default is dev")
	ddlCmd.Flags().BoolVarP(&reverse, "reverse", "r", false, "If true, generate entity code from database. If false, update or create database tables from entity code.")
	ddlCmd.Flags().BoolVarP(&dao, "dao", "d", false, "If true, generate dao code.")
}
