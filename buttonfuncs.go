package main

func exitApp() {
	logEvent("EXIT")
	defer App.Quit()
}

func stopTimer() {
	toggleStop()
}

func pauseTimer() {
	togglePause()
}
