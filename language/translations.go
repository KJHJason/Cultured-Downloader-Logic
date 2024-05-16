package language

// translations are done here and will be stored in the database

func initForDownloadQueue() {
	addData("date/time", "Date/Time", EN)
	addData("date/time", "日付/時間", JP)
	addData("your inputs", "Your Inputs", EN)
	addData("your inputs", "あなたの入力", JP)
	addData("inputs", "Inputs", EN)
	addData("inputs", "入力", JP)
	addData("progress", "Progress", EN)
	addData("actions", "Actions", EN)
	addData("actions", "アクション", JP)
	addData("current task", "Current Task", EN)
	addData("current task", "現在のタスク", JP)
	addData("downloads", "Downloads", EN)
	addData("downloads", "ダウンロード", JP)
}

func initForGeneral() {
	addData("save", "Save", EN)
	addData("save", "保存", JP)
	addData("unknown", "Unknown", EN)
	addData("unknown", "不明", JP)
	addData("settings", "Settings", EN)
	addData("settings", "設定", JP)
	addData("light mode", "Light Mode", EN)
	addData("light mode", "ライトモード", JP)
	addData("dark mode", "Dark Mode", EN)
	addData("dark mode", "ダークモード", JP)
}

func initForHomePage() {
	addData(
		"to get started, click on one of the options below or use the navigation bar in the top-left corner.",
		"To get started, click on one of the options below or use the navigation bar in the top-left corner.",
		EN,
	)
	addData(
		"to get started, click on one of the options below or use the navigation bar in the top-left corner.",
		"始めるには、以下のオプションのいずれかをクリックするか、左上隅のナビゲーションバーを使用してください。",
		JP,
	)
	addData("welcome back,", "Welcome back,", EN)
	addData("welcome back,", "おかえりなさい、", JP)
	addData("!", "!", EN)
	addData("!", "！", JP)
	addData("Home", "Home", EN)
	addData("Home", "ホーム", JP)
	addData("image:", "Image:", EN)
	addData("image:", "イラスト：", JP)
	addData("karutamo", "Karutamo", EN)
	addData("karutamo", "かるたも", JP)
	addData("found an issue? click me!", "Found an issue? Click me!", EN)
	addData("found an issue? click me!", "問題を発見しましたか？私をクリックしてください！", JP)
}

func initForProgramInfo() {
	addData("check for updates", "Check for Updates", EN)
	addData("check for updates", "更新を確認", JP)
	addData("checking for updates...", "Checking for updates...", EN)
	addData("checking for updates...", "更新を確認中...", JP)
	addData("outdated, last checked at", "Outdated, last checked at", EN)
	addData("outdated, last checked at", "古い、最後に確認した日時", JP)
	addData("up-to-date, last checked at", "Up-to-date, last checked at", EN)
	addData("up-to-date, last checked at", "最新、最後に確認した日時", JP)
}

func initForPagination() {
	addData("previous", "Previous", EN)
	addData("previous", "前", JP)
	addData("next", "Next", EN)
	addData("next", "次", JP)
	addData("showing", "Showing", EN)
	addData("showing", "エントリーの表示:", JP)
	addData("to", "to", EN)
	addData("to", "から", JP)
	addData("of", "of", EN)
	addData("of", "までの", JP)
	addData("entries", "entries", EN)
	addData("entries", "エントリー", JP)
}
