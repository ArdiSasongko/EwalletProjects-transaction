package api

func SetupHTTP() {
	app, err := SetupHTTPApplication()
	if err != nil {
		app.config.logger.Fatalf("failed setup application (http): %v", err)
	}

	api := app.mount()
	if err := app.run(api); err != nil {
		app.config.logger.Fatalf("failed to start http server: %v", err)
	}
}
