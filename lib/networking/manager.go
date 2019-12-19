package networking

func NewManager(config []*NetworkConfig) (*Manager, error) { //use the config to build the manager object and check all bridges etc..
  mgr := &Manager{
    bridges: map[string]NetworkBridge{},
  }
  if err := mgr.init(config); err != nil {
    return nil, err
  }
  return mgr, nil
}

type Manager struct {
  bridges map[string]NetworkBridge
  config []*NetworkConfig
}

func (mgr *Manager) init(config []*NetworkConfig) (error) {
  //initialise all configured bridges...
  mgr.config = config
  for _, brConfig := range config {
    if err := mgr.initBridge(brConfig); err != nil {
      return err
    }
  }
  return nil
}

func (mgr *Manager) initBridge(config *NetworkConfig) (error) {
  //initialise bridge...
  switch config.Type {
  case LinuxBridgeDriver:
    br, err := NewLinuxBridge(config)
    if err != nil {
      return err
    }
    mgr.bridges[config.ID] = br
  }
  return nil
}

func (mgr *Manager) CreateBridge(bridgeType BridgeDriver) (NetworkBridge, error) {
  return nil, nil
}

func (mgr *Manager) GetBridge(id string) (NetworkBridge, error) {
  return nil, nil
}

func (mgr *Manager) DestroyBridge(id string) (error) {
  return nil
}

func (mgr *Manager) cleanup() {
  //cleanup the bridges (teardown)
}
