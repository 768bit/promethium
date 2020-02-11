package lib

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/768bit/promethium/api/restapi"
	"github.com/768bit/promethium/api/restapi/operations"
	"github.com/768bit/promethium/lib/config"
	"github.com/768bit/promethium/lib/service"
	"github.com/768bit/promethium/lib/vmm"
	"github.com/go-openapi/loads"
	"github.com/gorilla/mux"
	"github.com/jessevdk/go-flags"
)

type PromethiumDaemonStatus uint8

const (
	DaemonStopped  PromethiumDaemonStatus = 0x00
	DaemonStopping PromethiumDaemonStatus = 0x01
	DaemonStarting PromethiumDaemonStatus = 0x02
	DaemonStarted  PromethiumDaemonStatus = 0x03
)

/* the daemon starts an instance of promethium - this can be managed via command line tools (which uses the API */

func NewPromethiumDaemon(foreground bool) (*PromethiumDaemon, error) {
	pd := &PromethiumDaemon{}
	if err := pd.init(foreground); err != nil {
		return nil, err
	}
	return pd, nil
}

type PromethiumDaemon struct {
	config     *config.PromethiumDaemonConfig
	exitChan   chan os.Signal
	waitChan   chan bool
	vmmManager *vmm.VmmManager
	status     PromethiumDaemonStatus
	api        *mux.Router
}

func (pd *PromethiumDaemon) init(foreground bool) error {

	//load the config...
	cfg, err := config.LoadPromethiumDaemonConfig()
	if err != nil {
		return err
	}
	pd.config = cfg
	// pd.storageManager, err := storage.NewStorageManager(cfg.Storage)
	// if err != nil {
	//   return err
	// }
	//pd.imgManager = images.NewImageManager(filepath.Join(pd.config.AppRoot, "images"))
	//capture interrupts/signals
	if !foreground {
		pd.setupDaemonise()
	} else {
		pd.setupForeground()
	}

	return nil

}

func (pd *PromethiumDaemon) captureInterrupts() {

	pd.exitChan = make(chan os.Signal)
	signal.Notify(pd.exitChan, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		for {
			select {
			case sig := <-pd.exitChan:
				switch sig {
				case syscall.SIGTERM:
					fmt.Println("SIGTERM")
					//terminate process... this has a timeout...
					pd.waitKillVmmManager()
					pd.waitChan <- true
				case os.Interrupt:
					fmt.Println("SIGINT")
					//interrupt process - no time out.. ctrl + c basically...
					pd.waitKillVmmManager()
					pd.waitChan <- true
				case syscall.SIGKILL:
					fmt.Println("SIGKILL")
					//forced quit - abort basically..
					pd.killVmmManager()
					pd.waitChan <- true
				}
			}
		}
	}()

}
func (pd *PromethiumDaemon) killVmmManager() error {
	log.Printf("Killing VmmManager and instances...")
	return pd.vmmManager.Kill()
}

func (pd *PromethiumDaemon) waitKillVmmManager() error {
	log.Printf("Killing VmmManager and instances with timeout...")
	return pd.vmmManager.WaitKill()
}

func (pd *PromethiumDaemon) setupDaemonise() {

	//start the daemon - which will start the api...

	//easyconfig.ParseFatal(nil, nil)

	service.Main(&service.Info{
		Title:       "Promethium Server Daemon",
		Name:        "promethium",
		Description: "The Promethium API Server Daemon",

		RunFunc: func(smgr service.Manager) error {
			// Start up your service.
			// ...

			log.Printf("Daemon Start...")

			// Once initialization requiring root is done, call this.
			if err := pd.Start(); err != nil {
				return err
			}

			log.Printf("Dropping Daemon Privs...")

			//if err := smgr.DropPrivileges(); err != nil {
			//	return err
			//	}

			log.Printf("Daemon Privs Dropped...")

			err := pd.vmmManager.Start()
			if err != nil {
				return err
			}

			// When it is ready to serve requests, call this.
			// You must call DropPrivileges first.
			smgr.SetStarted()

			// Optionally set a status.
			smgr.SetStatus("promethium: running ok")

			// Wait until stop is requested.
			<-smgr.StopChan()

			// Do any necessary teardown.
			// ...

			pd.waitKillVmmManager()

			log.Printf("Daemon Stop...")

			// Done.
			return nil
		},
	})

}

func (pd *PromethiumDaemon) setupForeground() {

	pd.waitChan = make(chan bool)

	pd.captureInterrupts()
	err := pd.Start()
	if err != nil {
		println(err.Error())
	}
	err = pd.vmmManager.Start()
	if err != nil {
		println(err.Error())
	}

}

func (pd *PromethiumDaemon) Start() error {

	//start the daemon - which will start the api...
	var err error
	if pd.vmmManager, err = vmm.NewVmmManager(pd.config); err != nil {
		return err
	}
	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		log.Fatalln(err)
	}
	api := operations.NewServerAPI(swaggerSpec)

	server := restapi.NewServer(api)
	restapi.SetManager(pd.vmmManager)
	server.ConfigureFlags()
	server.ConfigureAPI()

	if pd.config.Http != nil && pd.config.Http.Enable {
		server.EnabledListeners = append(server.EnabledListeners, "http")
		server.Host = pd.config.Http.BindAddress
		server.Port = int(pd.config.Http.Port)
	}

	if pd.config.Unix != nil && pd.config.Unix.Enable {
		server.EnabledListeners = append(server.EnabledListeners, "unix")
		if pd.config.Unix.Path == "" {
			server.SocketPath = config.PROMETHIUM_SOCKET_PATH
		} else {
			server.SocketPath = flags.Filename(pd.config.Unix.Path)
		}
		os.RemoveAll(string(server.SocketPath))
	}

	if pd.config.Https != nil && pd.config.Https.Enable {

	}

	go func() {
		println("Serving")
		err := server.Serve()
		if err != nil {
			println(err.Error())
		}
	}()

	//start api...
	// pd.api = api.MakeNewApiRouter(pd.vmmManager)
	// server := http.Server{
	// 	Handler: pd.api,
	// }
	// os.RemoveAll("/tmp/promethium")
	// unixListener, err := net.Listen("unix", "/tmp/promethium")
	// go func() {
	// 	log.Printf("Listening on Unix Socket...")
	// 	server.Serve(unixListener)
	// }()
	// go func() {
	// 	addr := fmt.Sprintf("%s:%d", pd.config.API.BindAddress, pd.config.API.Port)
	// 	log.Printf("Listening on %s...", addr)
	// 	http.ListenAndServe(addr, pd.api)
	// }()

	return nil

}

func (pd *PromethiumDaemon) stop() {

	//signal a shutdown to teardown everything...

}

func (pd *PromethiumDaemon) Wait() {

	//when running foreground wait for the system to shutdown...

	<-pd.waitChan

}
