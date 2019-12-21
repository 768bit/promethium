// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"bufio"
	"crypto/tls"
	"net/http"

	errors "github.com/go-openapi/errors"
	runtime "github.com/go-openapi/runtime"
	middleware "github.com/go-openapi/runtime/middleware"

	"github.com/768bit/promethium/api/models"
	"github.com/768bit/promethium/api/restapi/operations"
	"github.com/768bit/promethium/api/restapi/operations/images"
	"github.com/768bit/promethium/api/restapi/operations/networking"
	"github.com/768bit/promethium/api/restapi/operations/storage"
	"github.com/768bit/promethium/api/restapi/operations/vms"
	"github.com/768bit/promethium/lib/vmm"

	"github.com/gorilla/websocket"
)

var vmmManager *vmm.VmmManager

func SetManager(vmm *vmm.VmmManager) {
	vmmManager = vmm
}

//go:generate swagger generate server --target ../../api --name Server --spec ../swagger.yml --skip-models --exclude-main

func configureFlags(api *operations.ServerAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.ServerAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	if api.StorageGetStorageStorageIDDisksHandler == nil {
		api.StorageGetStorageStorageIDDisksHandler = storage.GetStorageStorageIDDisksHandlerFunc(func(params storage.GetStorageStorageIDDisksParams) middleware.Responder {
			return middleware.NotImplemented("operation storage.GetStorageStorageIDDisks has not yet been implemented")
		})
	}
	if api.StorageGetStorageStorageIDImagesHandler == nil {
		api.StorageGetStorageStorageIDImagesHandler = storage.GetStorageStorageIDImagesHandlerFunc(func(params storage.GetStorageStorageIDImagesParams) middleware.Responder {
			return middleware.NotImplemented("operation storage.GetStorageStorageIDImages has not yet been implemented")
		})
	}
	if api.StorageGetStorageStorageIDKernelsHandler == nil {
		api.StorageGetStorageStorageIDKernelsHandler = storage.GetStorageStorageIDKernelsHandlerFunc(func(params storage.GetStorageStorageIDKernelsParams) middleware.Responder {
			return middleware.NotImplemented("operation storage.GetStorageStorageIDKernels has not yet been implemented")
		})
	}
	if api.VmsCreateImageHandler == nil {
		api.VmsCreateImageHandler = vms.CreateImageHandlerFunc(func(params vms.CreateImageParams) middleware.Responder {
			return middleware.NotImplemented("operation vms.CreateImage has not yet been implemented")
		})
	}
	if api.NetworkingCreateNetworkHandler == nil {
		api.NetworkingCreateNetworkHandler = networking.CreateNetworkHandlerFunc(func(params networking.CreateNetworkParams) middleware.Responder {
			return middleware.NotImplemented("operation networking.CreateNetwork has not yet been implemented")
		})
	}
	if api.StorageCreateStorageHandler == nil {
		api.StorageCreateStorageHandler = storage.CreateStorageHandlerFunc(func(params storage.CreateStorageParams) middleware.Responder {
			return middleware.NotImplemented("operation storage.CreateStorage has not yet been implemented")
		})
	}
	api.VmsCreateVMHandler = vms.CreateVMHandlerFunc(func(params vms.CreateVMParams) middleware.Responder {
		newVm, err := vmmManager.Create(params.VMConfig)
		if err != nil {
			println(err.Error())
			return &vms.CreateVMOK{}
		}
		println(newVm.ID())
		return &vms.CreateVMOK{}
	})

	if api.VmsCreateVMDiskHandler == nil {
		api.VmsCreateVMDiskHandler = vms.CreateVMDiskHandlerFunc(func(params vms.CreateVMDiskParams) middleware.Responder {
			return middleware.NotImplemented("operation vms.CreateVMDisk has not yet been implemented")
		})
	}
	if api.VmsCreateVMInterfaceHandler == nil {
		api.VmsCreateVMInterfaceHandler = vms.CreateVMInterfaceHandlerFunc(func(params vms.CreateVMInterfaceParams) middleware.Responder {
			return middleware.NotImplemented("operation vms.CreateVMInterface has not yet been implemented")
		})
	}
	if api.VmsCreateVMVolumeHandler == nil {
		api.VmsCreateVMVolumeHandler = vms.CreateVMVolumeHandlerFunc(func(params vms.CreateVMVolumeParams) middleware.Responder {
			return middleware.NotImplemented("operation vms.CreateVMVolume has not yet been implemented")
		})
	}
	if api.VmsDeleteVMHandler == nil {
		api.VmsDeleteVMHandler = vms.DeleteVMHandlerFunc(func(params vms.DeleteVMParams) middleware.Responder {
			return middleware.NotImplemented("operation vms.DeleteVM has not yet been implemented")
		})
	}
	if api.VmsDeleteVMDriveHandler == nil {
		api.VmsDeleteVMDriveHandler = vms.DeleteVMDriveHandlerFunc(func(params vms.DeleteVMDriveParams) middleware.Responder {
			return middleware.NotImplemented("operation vms.DeleteVMDrive has not yet been implemented")
		})
	}
	if api.VmsDeleteVMInterfaceHandler == nil {
		api.VmsDeleteVMInterfaceHandler = vms.DeleteVMInterfaceHandlerFunc(func(params vms.DeleteVMInterfaceParams) middleware.Responder {
			return middleware.NotImplemented("operation vms.DeleteVMInterface has not yet been implemented")
		})
	}
	if api.VmsDeleteVMVolumeHandler == nil {
		api.VmsDeleteVMVolumeHandler = vms.DeleteVMVolumeHandlerFunc(func(params vms.DeleteVMVolumeParams) middleware.Responder {
			return middleware.NotImplemented("operation vms.DeleteVMVolume has not yet been implemented")
		})
	}
	if api.NetworkingDestroyNetworkHandler == nil {
		api.NetworkingDestroyNetworkHandler = networking.DestroyNetworkHandlerFunc(func(params networking.DestroyNetworkParams) middleware.Responder {
			return middleware.NotImplemented("operation networking.DestroyNetwork has not yet been implemented")
		})
	}
	if api.StorageDestroyStorageHandler == nil {
		api.StorageDestroyStorageHandler = storage.DestroyStorageHandlerFunc(func(params storage.DestroyStorageParams) middleware.Responder {
			return middleware.NotImplemented("operation storage.DestroyStorage has not yet been implemented")
		})
	}
	api.ImagesGetImagesListHandler = images.GetImagesListHandlerFunc(func(params images.GetImagesListParams) middleware.Responder {
		ls := vmmManager.Storage().GetImages()
		resp := &images.GetImagesListOK{
			Payload: make([]*models.Image, len(ls)),
		}
		for i, img := range ls {
			resp.Payload[i] = &img.Image
		}
		return resp
	})

	api.VmsStartVMHandler = vms.StartVMHandlerFunc(func(params vms.StartVMParams) middleware.Responder {
		vmm, err := vmmManager.Get(params.VMID)
		if err != nil {
			return &vms.StartVMNotFound{}
		}
		err = vmm.Start()
		if err != nil {
			println(err.Error())
			return &vms.StartVMBadRequest{}
		}
		return &vms.StartVMOK{}
	})

	if api.NetworkingGetNetworkHandler == nil {
		api.NetworkingGetNetworkHandler = networking.GetNetworkHandlerFunc(func(params networking.GetNetworkParams) middleware.Responder {
			return middleware.NotImplemented("operation networking.GetNetwork has not yet been implemented")
		})
	}
	if api.NetworkingGetNetworkInterfacesHandler == nil {
		api.NetworkingGetNetworkInterfacesHandler = networking.GetNetworkInterfacesHandlerFunc(func(params networking.GetNetworkInterfacesParams) middleware.Responder {
			return middleware.NotImplemented("operation networking.GetNetworkInterfaces has not yet been implemented")
		})
	}
	if api.NetworkingGetNetworkListHandler == nil {
		api.NetworkingGetNetworkListHandler = networking.GetNetworkListHandlerFunc(func(params networking.GetNetworkListParams) middleware.Responder {
			return middleware.NotImplemented("operation networking.GetNetworkList has not yet been implemented")
		})
	}
	if api.NetworkingGetPhysicalInterfacesHandler == nil {
		api.NetworkingGetPhysicalInterfacesHandler = networking.GetPhysicalInterfacesHandlerFunc(func(params networking.GetPhysicalInterfacesParams) middleware.Responder {
			return middleware.NotImplemented("operation networking.GetPhysicalInterfaces has not yet been implemented")
		})
	}
	if api.StorageGetStorageHandler == nil {
		api.StorageGetStorageHandler = storage.GetStorageHandlerFunc(func(params storage.GetStorageParams) middleware.Responder {
			return middleware.NotImplemented("operation storage.GetStorage has not yet been implemented")
		})
	}
	if api.StorageGetStorageListHandler == nil {
		api.StorageGetStorageListHandler = storage.GetStorageListHandlerFunc(func(params storage.GetStorageListParams) middleware.Responder {
			return middleware.NotImplemented("operation storage.GetStorageList has not yet been implemented")
		})
	}
	if api.VmsGetVMHandler == nil {
		api.VmsGetVMHandler = vms.GetVMHandlerFunc(func(params vms.GetVMParams) middleware.Responder {
			return middleware.NotImplemented("operation vms.GetVM has not yet been implemented")
		})
	}
	if api.VmsGetVMDiskHandler == nil {
		api.VmsGetVMDiskHandler = vms.GetVMDiskHandlerFunc(func(params vms.GetVMDiskParams) middleware.Responder {
			return middleware.NotImplemented("operation vms.GetVMDisk has not yet been implemented")
		})
	}
	if api.VmsGetVMDiskListHandler == nil {
		api.VmsGetVMDiskListHandler = vms.GetVMDiskListHandlerFunc(func(params vms.GetVMDiskListParams) middleware.Responder {
			return middleware.NotImplemented("operation vms.GetVMDiskList has not yet been implemented")
		})
	}
	if api.VmsGetVMInteraceHandler == nil {
		api.VmsGetVMInteraceHandler = vms.GetVMInteraceHandlerFunc(func(params vms.GetVMInteraceParams) middleware.Responder {
			return middleware.NotImplemented("operation vms.GetVMInterace has not yet been implemented")
		})
	}
	if api.VmsGetVMInterfaceListHandler == nil {
		api.VmsGetVMInterfaceListHandler = vms.GetVMInterfaceListHandlerFunc(func(params vms.GetVMInterfaceListParams) middleware.Responder {
			return middleware.NotImplemented("operation vms.GetVMInterfaceList has not yet been implemented")
		})
	}
	if api.VmsGetVMListHandler == nil {
		api.VmsGetVMListHandler = vms.GetVMListHandlerFunc(func(params vms.GetVMListParams) middleware.Responder {
			return middleware.NotImplemented("operation vms.GetVMList has not yet been implemented")
		})
	}
	if api.VmsGetVMVolumeHandler == nil {
		api.VmsGetVMVolumeHandler = vms.GetVMVolumeHandlerFunc(func(params vms.GetVMVolumeParams) middleware.Responder {
			return middleware.NotImplemented("operation vms.GetVMVolume has not yet been implemented")
		})
	}
	if api.VmsGetVMVolumeListHandler == nil {
		api.VmsGetVMVolumeListHandler = vms.GetVMVolumeListHandlerFunc(func(params vms.GetVMVolumeListParams) middleware.Responder {
			return middleware.NotImplemented("operation vms.GetVMVolumeList has not yet been implemented")
		})
	}
	if api.NetworkingUpdateNetworkHandler == nil {
		api.NetworkingUpdateNetworkHandler = networking.UpdateNetworkHandlerFunc(func(params networking.UpdateNetworkParams) middleware.Responder {
			return middleware.NotImplemented("operation networking.UpdateNetwork has not yet been implemented")
		})
	}
	if api.StorageUpdateStorageHandler == nil {
		api.StorageUpdateStorageHandler = storage.UpdateStorageHandlerFunc(func(params storage.UpdateStorageParams) middleware.Responder {
			return middleware.NotImplemented("operation storage.UpdateStorage has not yet been implemented")
		})
	}
	if api.VmsUpdateVMHandler == nil {
		api.VmsUpdateVMHandler = vms.UpdateVMHandlerFunc(func(params vms.UpdateVMParams) middleware.Responder {
			return middleware.NotImplemented("operation vms.UpdateVM has not yet been implemented")
		})
	}
	if api.VmsUpdateVMDiskHandler == nil {
		api.VmsUpdateVMDiskHandler = vms.UpdateVMDiskHandlerFunc(func(params vms.UpdateVMDiskParams) middleware.Responder {
			return middleware.NotImplemented("operation vms.UpdateVMDisk has not yet been implemented")
		})
	}
	if api.VmsUpdateVMInterfaceHandler == nil {
		api.VmsUpdateVMInterfaceHandler = vms.UpdateVMInterfaceHandlerFunc(func(params vms.UpdateVMInterfaceParams) middleware.Responder {
			return middleware.NotImplemented("operation vms.UpdateVMInterface has not yet been implemented")
		})
	}
	if api.VmsUpdateVMVolumeHandler == nil {
		api.VmsUpdateVMVolumeHandler = vms.UpdateVMVolumeHandlerFunc(func(params vms.UpdateVMVolumeParams) middleware.Responder {
			return middleware.NotImplemented("operation vms.UpdateVMVolume has not yet been implemented")
		})
	}

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return wsMiddleware(handler)
}

func wsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Our middleware logic goes here...
		if r.URL.Path == "/consolews" {
			//upgrade the connection as needed...
			serveWs(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

var upgrader = websocket.Upgrader{}

type InboundJsonMessage struct {
	ID        string                 `json:"id"`
	Operation string                 `json:"operation"`
	Payload   map[string]interface{} `json:"payload"`
}

type OutboundJsonMessage struct {
	ID        string                 `json:"id"`
	Operation string                 `json:"operation"`
	Payload   map[string]interface{} `json:"payload"`
	Code      int                    `json:"code"`
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		println("upgrade:", err)
		return
	}

	for {
		inboundMsg := &InboundJsonMessage{}
		err := ws.ReadJSON(inboundMsg)
		if err != nil {
			println("read:", err.Error())
			break
		}
		switch inboundMsg.Operation {
		case "connect-console":
			vmID := inboundMsg.Payload["id"].(string)
			err = handleConsoleConnect(ws, vmID)
			if err != nil {
				println(err.Error())
			}
		}
	}

	defer ws.Close()

}

func handleConsoleConnect(ws *websocket.Conn, id string) error {
	vmm, err := vmmManager.Get(id)
	if err != nil {
		return err
	} else {
		outP, errP, inP, err := vmm.Console()
		if err != nil {
			return err
		}
		go func() {
			//use the pipes..
			outScanner := bufio.NewReader(outP)
			obuff := make([]byte, 1024)
			for {

				n, err := outScanner.Read(obuff)
				if err != nil {
					println("read:", err.Error())
					return
				} else if n == 0 {
					continue
				}

				err = ws.WriteJSON(&OutboundJsonMessage{
					ID:        "",
					Operation: "console-output",
					Code:      0,
					Payload: map[string]interface{}{
						"output": string(obuff[:n]),
					},
				})
				if err != nil {
					println("write:", err.Error())
					return
				}
			}
		}()
		go func() {
			//use the pipes..
			outScanner := bufio.NewReader(errP)
			obuff := make([]byte, 1024)
			for {

				n, err := outScanner.Read(obuff)
				if err != nil {
					println("read:", err.Error())
					return
				} else if n == 0 {
					continue
				}

				err = ws.WriteJSON(&OutboundJsonMessage{
					ID:        "",
					Operation: "console-output",
					Code:      0,
					Payload: map[string]interface{}{
						"output": string(obuff[:n]),
					},
				})
				if err != nil {
					println("write:", err.Error())
					return
				}
			}
		}()
		for {
			inboundMsg := &InboundJsonMessage{}
			err := ws.ReadJSON(inboundMsg)
			if err != nil {
				println("read:", err.Error())
				break
			}
			switch inboundMsg.Operation {
			case "console-input":
				inp := inboundMsg.Payload["input"].(string)
				inP.Write([]byte(inp))
			default:
				break
			}
		}
	}
	return nil
}
