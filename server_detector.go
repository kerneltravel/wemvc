package wemvc

type configDetector struct {
	app *server
}

func (d *configDetector) CanHandle(path string) bool {
	return d.app.isConfigFile(path)
}

type nsConfigDetector struct {
	app *server
}

func (d *nsConfigDetector) CanHandle(path string) bool {
	return d.app.isNsConfigFile(path)
}

type viewDetector struct {
	app *server
}

func (d *viewDetector) CanHandle(path string) bool {
	return d.app.isInViewFolder(path)
}

type nsViewDetector struct {
	app *server
}

func (d *nsViewDetector) CanHandle(path string) bool {
	for _, ns := range d.app.namespaces {
		if ns.isInViewFolder(path) {
			return true
		}
	}
	return false
}