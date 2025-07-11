package types

type UploadForm struct {
	Process  string `json:"process"`
	Title    string `json:"title"`
	Time     string `json:"time"`
	Username string `json:"user_name"`
}

type UploadResult struct {
	Code int              `json:"code"`
	Data UploadDataResult `json:"data"`
	Msg  string           `json:"msg"`
}

type UploadDataResult struct {
	Lave   int    `json:"lave"`
	Notice string `json:"notice"`
}
