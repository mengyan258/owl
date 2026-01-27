package owl

import (
	"embed"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/kardianos/service"
	"github.com/spf13/cobra"
	"gopkg.in/guoliang1994/go-i18n.v2"
	"gopkg.in/guoliang1994/go-i18n.v2/driver"
)

type program struct {
	binName     string
	description string
	version     string
	program     service.Interface
	service     service.Service
	rootCmd     *cobra.Command
	lang        *i18n.I18N
}

func (p *program) AddCommand(cmd ...*cobra.Command) *program {
	p.rootCmd.AddCommand(cmd...)
	return p
}

type installer struct {
	programs       []*program
	rootCmd        *cobra.Command
	programLen     int
	hasRootProgram bool
}

func NewInstaller() *installer {
	return &installer{}
}

func (i *installer) RootProgram(binName, description string) *installer {
	i.rootCmd = &cobra.Command{
		Use:               binName,
		Short:             description,
		Long:              description,
		CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
	}
	i.hasRootProgram = true
	return i
}
func (i *installer) AddProgram(p *program) *installer {
	i.programs = append(i.programs, p)
	i.programLen++
	return i
}
func (i *installer) execute() {
	for _, p := range i.programs {
		p.start()
		p.stop()
		p.restart()
		p.install()
		p.uninstall()
		p.status()
		p.ver()
		p.run()
		p.Lang()
	}
	cobra.CheckErr(i.rootCmd.Execute())
}
func (i *installer) isMultiInstall() bool {
	return i.programLen > 1
}
func (i *installer) Install() {
	for _, p := range i.programs {
		// 初始化 service
		options := make(service.KeyValue)
		svcConfig := &service.Config{
			Name:        p.binName,
			DisplayName: p.binName,
			Description: p.description,
			Option:      options,
		}

		if runtime.GOOS != "windows" {
			svcConfig.Dependencies = []string{
				"Requires=network.target",
				"After=network-online.target syslog.target"}
			svcConfig.UserName = "root"
		}
		var err error
		// 增加 service 运行时的参数，最后得到 nps run
		if i.isMultiInstall() {
			if i.hasRootProgram {
				// 如果有多个程序，则需要增加程序名称再 run
				svcConfig.Arguments = append(svcConfig.Arguments, p.binName)
				i.rootCmd.AddCommand(p.rootCmd)
			} else {
				i.rootCmd = p.rootCmd
				fmt.Println(p.lang.T("installMultiProgram"), p.binName)
			}
		} else {
			// 如果只有一个程序，那根程序就等于子程序
			i.rootCmd = p.rootCmd
		}
		svcConfig.Arguments = append(svcConfig.Arguments, "run")
		p.service, err = service.New(p.program, svcConfig)
		if err != nil {
			fmt.Println(err)
		}
	}
	i.execute()
}

//go:embed lang
var langFs embed.FS

func NewProgram(binName, description, version string, p service.Interface) *program {

	if err := IsValidBinaryName(binName); err != nil {
		panic("Invalid binary name: " + err.Error())
	}

	// Use Chinese as the default language
	// Modify the language using the Lang command
	embedDriver := driver.NewEmbedI18NImpl(langFs, "lang/")
	_, err := os.Stat("lang.conf")
	var l *i18n.I18N
	if err != nil {
		l = i18n.NewI18N(i18n.Chinese)
	} else {
		lang, _ := ioutil.ReadFile("lang.conf")
		l = i18n.NewI18N(string(lang))
	}
	l.AddLang(embedDriver)
	app := &program{
		binName:     binName,
		description: description,
		version:     version,
		program:     p,
		lang:        l,
	}
	// 程序的名称就是根命令
	app.rootCmd = &cobra.Command{
		Use:  binName,
		Long: description,
	}
	// 隐藏默认的命令
	app.rootCmd.CompletionOptions.HiddenDefaultCmd = true
	return app
}

func (i *program) install() {
	use := "install"
	var c = &cobra.Command{
		Use:   use,
		Short: i.lang.T("install.short", i.binName),
		Long:  i.lang.T("install.long", i.binName),
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = i.service.Stop()
			_ = i.service.Uninstall()
			err := i.service.Install()
			if err != nil {
				return errors.New(i.lang.T("install.fail", i.binName, err.Error()))
			}
			fmt.Println(i.lang.T("install.success", i.binName))
			return nil
		},
	}
	i.rootCmd.AddCommand(c)
}

func (i *program) uninstall() {
	use := "uninstall"
	var c = &cobra.Command{
		Use:   use,
		Short: i.lang.T("uninstall.short", i.binName),
		Long:  i.lang.T("uninstall.short", i.binName),
		RunE:  i.control(use),
	}
	i.rootCmd.AddCommand(c)
}

func (i *program) run() {
	use := "run"
	var c = &cobra.Command{
		Use:   use,
		Short: i.lang.T("foreground.short", i.binName),
		Long:  i.lang.T("foreground.short", i.binName),
		RunE: func(cmd *cobra.Command, args []string) error {
			return i.service.Run()
		},
	}
	i.rootCmd.AddCommand(c)
}
func (i *program) start() {
	use := "start"
	var c = &cobra.Command{
		Use:   use,
		Short: i.lang.T("start.short", i.binName),
		Long:  i.lang.T("start.short", i.binName),
		RunE:  i.control(use),
	}
	i.rootCmd.AddCommand(c)
}

func (i *program) stop() {
	use := "stop"
	var c = &cobra.Command{
		Use:   use,
		Short: i.lang.T("stop.short", i.binName),
		Long:  i.lang.T("stop.short", i.binName),
		RunE:  i.control(use),
	}
	i.rootCmd.AddCommand(c)
}
func (i *program) restart() {
	use := "restart"
	var c = &cobra.Command{
		Use:   use,
		Short: i.lang.T("restart.short", i.binName),
		Long:  i.lang.T("restart.long", i.binName),
		RunE:  i.control(use),
	}
	i.rootCmd.AddCommand(c)
}
func (i *program) status() {
	use := "status"
	var c = &cobra.Command{
		Use:   use,
		Short: i.lang.T("status.short", i.binName),
		Long:  i.lang.T("status.short", i.binName),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("")
		},
	}
	i.rootCmd.AddCommand(c)
}
func (i *program) ver() {
	use := "version"
	var c = &cobra.Command{
		Use:   use,
		Short: i.lang.T("version.short", i.binName),
		Long:  i.lang.T("version.short", i.binName),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(i.version)
		},
	}
	c.Flags()
	i.rootCmd.AddCommand(c)
}
func (i *program) Lang() {
	use := "lang"
	var langCmd = &cobra.Command{
		Use:   use,
		Short: i.lang.T("lang.short", i.binName),
		Long:  i.lang.T("lang.short", i.binName),
		Run: func(cmd *cobra.Command, args []string) {
			lang := cmd.Flag("language").Value.String()
			err := ioutil.WriteFile("lang.conf", []byte(lang), 0777)
			if err != nil {
				return
			}
		},
	}
	var lang string
	langCmd.Flags().StringVarP(&lang, "language", "l", "zh", i.lang.T("lang.short"))
	i.rootCmd.AddCommand(langCmd)
}

// start stop restart
func (i *program) control(command string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if service.Platform() == "unix-systemv" {
			terminal := exec.Command("/etc/init.d/"+i.binName, command)
			return terminal.Run()
		}
		err := service.Control(i.service, command)
		if err != nil {
			return errors.New(i.lang.T(command+".fail", i.binName, err.Error()))
		}
		return nil
	}
}

// IsValidBinaryName 检查二进制文件名是否合法
func IsValidBinaryName(binName string) error {
	// 检查是否为空
	if strings.TrimSpace(binName) == "" {
		return errors.New("binary name cannot be empty")
	}

	// 检查长度
	if len(binName) < 1 || len(binName) > 50 {
		return errors.New("binary name length must be between 1 and 50 characters")
	}

	// 使用正则表达式检查字符合法性
	// 允许字母、数字、下划线、连字符
	validPattern := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	if !validPattern.MatchString(binName) {
		return errors.New("binary name can only contain letters, digits, underscores, and hyphens, and must start with a letter")
	}

	// 检查是否包含路径分隔符（这会是危险的）
	if strings.Contains(binName, "/") || strings.Contains(binName, "\\") {
		return errors.New("binary name cannot contain path separators")
	}

	return nil
}
