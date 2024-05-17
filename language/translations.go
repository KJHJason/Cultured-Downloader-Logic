package language

// translations are done here and will be stored in the database

func initForDownloadQueue(db *dataInitWrapper) {
	db.addData("date/time", "Date/Time", EN)
	db.addData("date/time", "日付/時間", JP)
	db.addData("your inputs", "Your Inputs", EN)
	db.addData("your inputs", "あなたの入力", JP)
	db.addData("inputs", "Inputs", EN)
	db.addData("inputs", "入力", JP)
	db.addData("progress", "Progress", EN)
	db.addData("actions", "Actions", EN)
	db.addData("actions", "アクション", JP)
	db.addData("current task", "Current Task", EN)
	db.addData("current task", "現在のタスク", JP)
	db.addData("downloads", "Downloads", EN)
	db.addData("downloads", "ダウンロード", JP)
}

func initForGeneral(db *dataInitWrapper) {
	db.addData("save", "Save", EN)
	db.addData("save", "保存", JP)
	db.addData("unknown", "Unknown", EN)
	db.addData("unknown", "不明", JP)
	db.addData("settings", "Settings", EN)
	db.addData("settings", "設定", JP)
	db.addData("light mode", "Light Mode", EN)
	db.addData("light mode", "ライトモード", JP)
	db.addData("dark mode", "Dark Mode", EN)
	db.addData("dark mode", "ダークモード", JP)
}

func initForHomePage(db *dataInitWrapper) {
	db.addData(
		"to get started, click on one of the options below or use the navigation bar in the top-left corner.",
		"To get started, click on one of the options below or use the navigation bar in the top-left corner.",
		EN,
	)
	db.addData(
		"to get started, click on one of the options below or use the navigation bar in the top-left corner.",
		"始めるには、以下のオプションのいずれかをクリックするか、左上隅のナビゲーションバーを使用してください。",
		JP,
	)
	db.addData("welcome back,", "Welcome back,", EN)
	db.addData("welcome back,", "おかえりなさい、", JP)
	db.addData("!", "!", EN)
	db.addData("!", "！", JP)
	db.addData("home", "Home", EN)
	db.addData("home", "ホーム", JP)
	db.addData("image:", "Image:", EN)
	db.addData("image:", "イラスト：", JP)
	db.addData("karutamo", "Karutamo", EN)
	db.addData("karutamo", "かるたも", JP)
	db.addData("found an issue? click me!", "Found an issue? Click me!", EN)
	db.addData("found an issue? click me!", "問題を発見しましたか？私をクリックしてください！", JP)
}

func initForProgramInfo(db *dataInitWrapper) {
	db.addData("check for updates", "Check for Updates", EN)
	db.addData("check for updates", "更新を確認", JP)
	db.addData("checking for updates...", "Checking for updates...", EN)
	db.addData("checking for updates...", "更新を確認中...", JP)
	db.addData("outdated, last checked at", "Outdated, last checked at", EN)
	db.addData("outdated, last checked at", "古い、最後に確認した日時", JP)
	db.addData("up-to-date, last checked at", "Up-to-date, last checked at", EN)
	db.addData("up-to-date, last checked at", "最新、最後に確認した日時", JP)
}

func initForPagination(db *dataInitWrapper) {
	db.addData("previous", "Previous", EN)
	db.addData("previous", "前", JP)
	db.addData("next", "Next", EN)
	db.addData("next", "次", JP)
	db.addData("showing", "Showing", EN)
	db.addData("showing", "エントリーの表示:", JP)
	db.addData("to", "to", EN)
	db.addData("to", "から", JP)
	db.addData("of", "of", EN)
	db.addData("of", "までの", JP)
	db.addData("entries", "entries", EN)
	db.addData("entries", "エントリー", JP)
}
