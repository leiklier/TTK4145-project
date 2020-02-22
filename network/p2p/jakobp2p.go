package jakobp2p

func initialize() {
	if isInitialized {
		return
	}
	isInitialized = true
	localIP, _ = getLocalIP()
	go peersServer()
}
