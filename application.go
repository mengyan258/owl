package owl

import (
	"context"
	"fmt"
	"go/types"
	"os"
	"path/filepath"
	"reflect"
	"unsafe"

	"bit-labs.cn/owl/provider/appconf"
	"bit-labs.cn/owl/provider/validator"

	"bit-labs.cn/owl/contract/foundation"
	logContract "bit-labs.cn/owl/contract/log"
	"bit-labs.cn/owl/provider/conf"
	"bit-labs.cn/owl/provider/event"
	"bit-labs.cn/owl/provider/log"
	"bit-labs.cn/owl/provider/router"
	"bit-labs.cn/owl/utils"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
)

// SubApp 子应用
type SubApp interface {
	Name() string
	RegisterRouters()
	ServiceProviders() []foundation.ServiceProvider
	Binds() []any
	Menu() []*router.Menu
	Commands() []*cobra.Command
	// Bootstrap 应用启动前执行，如初始化配置,初始化数据，初始化表结构等
	Bootstrap()
}

const (
	version = "1.0.0"
	/**
	 * The environment file to load during bootstrapping.
	 *
	 * @var string
	 */
	environmentFile = ".env"
)

var (
	instance *Application
)

var _ foundation.Application = (*Application)(nil)

type Application struct {
	*dig.Container

	/**
	 * The base path for the owl installation.
	 *
	 * @var string
	 */
	basePath string

	bootingCallbacks []func(application foundation.Application)
	bootedCallbacks  []func(application foundation.Application)

	/**
	 * The array of terminating callbacks.
	 *
	 * @var callable[]
	 */
	terminatingCallbacks []types.Func

	/**
	 * All of the registered service providers.
	 *
	 * @var \Illuminate\Support\ServiceProvider[]
	 */
	serviceProviders []foundation.ServiceProvider

	/**
	 * The names of the loaded service providers.
	 *
	 * @var array
	 */
	loadedProviders []string

	/**
	 * The custom configuration path defined by the developer.
	 *
	 * @var string
	 */
	configPath string

	/**
	 * The custom database path defined by the developer.
	 *
	 * @var string
	 */
	databasePath string

	/**
	 * The custom language file path defined by the developer.
	 *
	 * @var string
	 */
	langPath string

	/**
	 * The custom public / web path defined by the developer.
	 *
	 * @var string
	 */
	publicPath string

	/**
	 * The custom storage path defined by the developer.
	 *
	 * @var string
	 */
	storagePath string

	/**
	 * The custom environment path defined by the developer.
	 *
	 * @var string
	 */
	environmentPath string

	/**
	 * Indicates if the application is running in the console.
	 *
	 * @var bool|null
	 */
	isRunningInConsole bool

	runDir          string
	name            string
	menuManager     *router.MenuRepository
	serviceProvider []foundation.ServiceProvider
	subApps         []SubApp
	rootCmd         *cobra.Command
	binds           []any
	menus           []*router.Menu

	l logContract.Logger
}

func (i *Application) Version() string {
	return version
}

func (i *Application) GetBasePath() string {
	return i.basePath
}

func (i *Application) GetBootstrapPath() string {
	return i.runDir
}

func (i *Application) GetConfigPath() string {
	return i.inferDir("conf")
}

func (i *Application) GetLangPath() string {
	return i.inferDir("lang")
}

func (i *Application) GetPublicPath() string {
	return i.inferDir("public")
}

func (i *Application) GetResourcePath() string {
	return i.inferDir("resource")
}

func (i *Application) GetStoragePath() string {
	return i.inferDir("storage")
}

func (i *Application) Environment(s ...string) (string, bool) {
	//TODO implement me
	panic("implement me")
}

func (i *Application) MaintenanceMode() context.Context {
	//TODO implement me
	panic("implement me")
}

func (i *Application) IsDownForMaintenance() bool {
	//TODO implement me
	panic("implement me")
}

func (i *Application) Register(providers ...any) {
	for _, p := range providers {
		err := i.Provide(p)
		if err != nil {
			err = i.Invoke(func(l logContract.Logger) {
				//l.Info(err.Error())
			})
			PanicIf(err)
		}
	}
}

func (i *Application) Provide(function interface{}, opts ...dig.ProvideOption) error {
	return i.Container.Provide(function, opts...)
}
func (i *Application) Invoke(function interface{}, opts ...dig.InvokeOption) error {
	return i.Container.Invoke(function, opts...)
}

func (i *Application) Boot() {
	//TODO implement me
	panic("implement me")
}

func (i *Application) Booting(callback func(application foundation.Application)) {
	i.bootingCallbacks = append(i.bootingCallbacks, callback)
}

func (i *Application) Booted(callback func(application foundation.Application)) {
	i.bootedCallbacks = append(i.bootedCallbacks, callback)
}

func (i *Application) GetLocale() string {
	//TODO implement me
	panic("implement me")
}

func (i *Application) GetProviders(provider interface{}) []interface{} {
	//TODO implement me
	panic("implement me")
}

func (i *Application) HasBeenBootstrapped() bool {
	//TODO implement me
	panic("implement me")
}

func (i *Application) SetLocale(locale string) {
	//TODO implement me
	panic("implement me")
}

func (i *Application) Terminating(callback interface{}) foundation.Application {
	//TODO implement me
	panic("implement me")
}

func (i *Application) Terminate() {
	//TODO implement me
	panic("implement me")
}

func NewApp(apps ...SubApp) *Application {
	i := &Application{
		rootCmd:   &cobra.Command{Use: "owl"},
		Container: dig.New(),
	}

	i.setPath()
	i.ensureConfDir()
	i.registerBaseBindings()
	i.registerBaseServiceProviders()
	i.newSubApp(apps...)
	return i
}

// 设置路径
func (i *Application) setPath() {
	// 获取启动当前程序的目录
	var err error
	i.runDir, err = os.Getwd()
	if err != nil {
		panic(err)
	}
	utils.PrintLnGreen(fmt.Sprintf("程序启动目录：%s ", i.runDir))

	// 获取当前可执行文件的路径
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}

	utils.PrintLnGreen(fmt.Sprintf("程序所在目录：%s", exePath))
	i.basePath = filepath.Dir(exePath)
}

func (i *Application) ensureConfDir() {
	dir := i.GetConfigPath()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		PanicIf(err)
	}
}
func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}
func (i *Application) Run() {
	err := i.Invoke(func(e *router.RouterServiceProvider) {
		e.Run()
	})
	PanicIf(err)
}

func (i *Application) ShowProviders() {
	i.l.Info("============================已注册服务提供者==========================================")
	for _, provider := range i.serviceProvider {
		providerType := reflect.TypeOf(provider).String()
		i.l.Info(providerType, provider.Description())
	}

	i.l.Info("======================================================================================")

}

// 注册基础服务提供者
func (i *Application) registerBaseServiceProviders() {
	var baseProviders = []foundation.ServiceProvider{
		&conf.ConfServiceProvider{},
		&log.LogServiceProvider{},
		&event.EventServiceProvider{},
		&router.RouterServiceProvider{},
		&appconf.AppConfigServiceProvider{},
		&validator.ValidatorServiceProvider{},
	}

	i.bootServiceProviders(baseProviders...)

	err := i.Invoke(func(l logContract.Logger) {
		i.l = l
	})

	PanicIf(err)
}

func (i *Application) bootServiceProviders(provider ...foundation.ServiceProvider) {

	for _, serviceProvider := range provider {
		i.injectAppInstance(serviceProvider)
		serviceProvider.Register()

		cfgFileGen := serviceProvider.Conf()
		if cfgFileGen != nil {
			for fileName, content := range cfgFileGen {
				confFile := i.GetConfigPath() + "/" + fileName
				_, err := os.Stat(confFile)
				if os.IsNotExist(err) {
					_ = os.WriteFile(confFile, []byte(content), 0644)
				}
			}
		}
	}
}

func (i *Application) newSubApp(apps ...SubApp) {
	for _, app := range apps {
		i.injectAppInstance(app)
		i.subApps = append(i.subApps, app)
	}

	for _, app := range i.subApps {
		i.binds = append(i.binds, app.Binds()...)
		i.serviceProvider = append(i.serviceProvider, app.ServiceProviders()...)
	}

	i.bootServiceProviders(i.serviceProvider...)

	for _, bind := range i.binds {
		err := i.Provide(bind)
		PanicIf(err)
	}

	for _, app := range i.subApps {
		app.RegisterRouters()
		i.menus = append(i.menus, app.Menu()...)
		i.rootCmd.AddCommand(app.Commands()...)
		app.Bootstrap()
	}

	// 将所有子应用的菜单添加到菜单管理器
	i.menuManager.AddMenu(i.menus...)

	for _, serviceProvider := range i.serviceProvider {
		serviceProvider.Boot()
	}

	i.ShowProviders()
}

func (i *Application) injectAppInstance(target any) {
	field := reflect.Indirect(reflect.ValueOf(target)).FieldByName("app")
	fieldPtr := unsafe.Pointer(field.UnsafeAddr())
	fieldValue := reflect.NewAt(field.Type(), fieldPtr).Elem()
	fieldValue.Set(reflect.ValueOf(i))
}

func (i *Application) registerBaseBindings() {
	err := i.Provide(func() *Application {
		return i
	}, dig.As(new(foundation.Application)))
	PanicIf(err)

	err = i.Provide(router.NewMenuRepository)
	PanicIf(err)
}

// inferDir 推断目录路径
func (i *Application) inferDir(path string) string {

	// 先尝试使用 basePath 拼接路径
	dir := filepath.Join(i.basePath, path)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// 如果 basePath 拼接路径不存在，则尝试使用 runDir 拼接路径
		dir = filepath.Join(i.runDir, path)
	}

	return dir
}
