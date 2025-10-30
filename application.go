package owl

import (
	"bit-labs.cn/owl/contract/foundation"
	logContract "bit-labs.cn/owl/contract/log"
	conf2 "bit-labs.cn/owl/provider/conf"
	"bit-labs.cn/owl/provider/event"
	"bit-labs.cn/owl/provider/log"
	"bit-labs.cn/owl/provider/router"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
	"go/types"
	"os"
	"path/filepath"
	"reflect"
	"unsafe"
)

// SubApp 子应用
type SubApp interface {
	Name() string
	RegisterRouters()
	ServiceProviders() []foundation.ServiceProvider
	Binds() []any
	Menu() *router.Menu
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
	/**
	 * Indicates if the application has been bootstrapped before.
	 *
	 * @var bool
	 */
	hasBeenBootstrapped bool
	/**
	 * Indicates if the application has "booted".
	 *
	 * @var bool
	 */
	booted bool

	/**
	 * The array of booting callbacks.
	 *
	 * @var callable[]
	 */
	bootingCallbacks []types.Func

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

func (i *Application) BasePath(path string) string {
	return i.basePath
}

func (i *Application) BootstrapPath(path string) string {
	return i.runDir
}

func (i *Application) ConfigPath(path string) string {
	if path == "" {
		path = "conf"
	}

	confDir := filepath.Join(i.basePath, path)
	// 如果 confDir 存在
	if _, err := os.Stat(confDir); os.IsNotExist(err) {
		confDir = filepath.Join(i.runDir, path)
	}

	return confDir
}

func (i *Application) DatabasePath(path string) string {
	//TODO implement me
	panic("implement me")
}

func (i *Application) LangPath(path string) string {
	return path
}

func (i *Application) PublicPath(path string) string {
	//TODO implement me
	panic("implement me")
}

func (i *Application) ResourcePath(path string) string {
	return ""
}

func (i *Application) StoragePath(path string) string {
	return i.basePath + "./storage"
}

func (i *Application) Environment(s ...string) (string, bool) {
	//TODO implement me
	panic("implement me")
}

func (i *Application) RunningInConsole() bool {
	//TODO implement me
	panic("implement me")
}

func (i *Application) RunningUnitTests() bool {
	//TODO implement me
	panic("implement me")
}

func (i *Application) HasDebugModeEnabled() bool {
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
				l.Info(err.Error())
			})
			PanicIf(err)
		}
	}
}

func (i *Application) RegisterDeferredProvider(provider string, service *string) {
	//TODO implement me
	panic("implement me")
}

func (i *Application) ResolveProvider(provider string) interface{} {
	panic("implement me")
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

func (i *Application) Booting(callback func()) {
	//TODO implement me
	panic("implement me")
}

func (i *Application) Booted(callback func()) {
	//TODO implement me
	panic("implement me")
}

func (i *Application) BootstrapWith(bootstrappers []string) {
	//TODO implement me
	panic("implement me")
}

func (i *Application) GetLocale() string {
	//TODO implement me
	panic("implement me")
}

func (i *Application) GetNamespace() string {
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

func (i *Application) LoadDeferredProviders() {
	//TODO implement me
	panic("implement me")
}

func (i *Application) SetLocale(locale string) {
	//TODO implement me
	panic("implement me")
}

func (i *Application) ShouldSkipMiddleware() bool {
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

	i.setBasePath()
	i.registerBaseBindings()
	i.registerBaseServiceProviders()

	i.newSubApp(apps...)
	return i
}

// 设置路径
func (i *Application) setBasePath() {
	var err error
	i.runDir, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	// 获取当前可执行文件的路径
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}

	i.basePath = filepath.Dir(exePath)
}
func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}
func (i *Application) Run() {
	err := i.Invoke(func(router *gin.Engine, configure *conf2.Configure, l logContract.Logger) {

		i.l = l
		var AppConfig struct {
			Listen string
		}
		err := configure.GetConfig("app", &AppConfig)
		if err != nil {
			return
		}

		_ = router.Run(AppConfig.Listen)
	})
	PanicIf(err)
}

func (i *Application) registerBaseServiceProviders() {
	var baseProviders = []foundation.ServiceProvider{
		&conf2.ConfServiceProvider{},
		&log.LogServiceProvider{},
		&event.EventServiceProvider{},
		&router.RouterServiceProvider{},
	}

	for _, serviceProvider := range baseProviders {
		i.injectAppInstance(serviceProvider)
		serviceProvider.Register()
	}
}

func (i *Application) RegisterServiceProviders(serviceProvider ...foundation.ServiceProvider) {
	i.serviceProvider = append(i.serviceProvider, serviceProvider...)
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

	for _, serviceProvider := range i.serviceProvider {
		i.injectAppInstance(serviceProvider)
		serviceProvider.Register()
	}

	for _, bind := range i.binds {
		err := i.Provide(bind)
		PanicIf(err)
	}

	for _, app := range i.subApps {
		app.RegisterRouters()
		i.menus = append(i.menus, app.Menu())
		i.rootCmd.AddCommand(app.Commands()...)
		app.Bootstrap()
	}

	// 将所有子应用的菜单添加到菜单管理器
	i.menuManager.AddMenu(i.menus...)

	for _, serviceProvider := range i.serviceProvider {
		serviceProvider.Boot()
	}
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
