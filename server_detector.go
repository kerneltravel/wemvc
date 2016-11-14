package wemvc

type configDetector struct {
	app *server
}

func (d *configDetector) CanHandle(path string) (bool, interface{}) {
	return d.app.isConfigFile(path), nil
}

type nsConfigDetector struct {
	app *server
}

func (d *nsConfigDetector) CanHandle(path string) (bool, interface{}) {
	for _, ns := range app.namespaces {
		if ns.isConfigFile(path) {
			return true, ns
		}
	}
	return false, nil
}

type viewDetector struct {
	app *server
}

func (d *viewDetector) CanHandle(path string) (bool, interface{}) {
	return d.app.isInViewFolder(path), nil
}

type nsViewDetector struct {
	app *server
}

func (d *nsViewDetector) CanHandle(path string) (bool,interface{}) {
	for _, ns := range d.app.namespaces {
		if ns.isInViewFolder(path) {
			return true, ns
		}
	}
	return false, nil
}