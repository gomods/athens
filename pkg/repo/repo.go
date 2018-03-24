package repo

type Crawler interface {
	// Downloads repo to tmp folder, path to tmp returned
	DownloadRepo() (string, error)
}
