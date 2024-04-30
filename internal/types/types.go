package types

type GetCarQuery struct {
	RegNum     string `in:"query=regNum"`
	Mark       string `in:"query=mark"`
	Model      string `in:"query=model"`
	Year       int    `in:"query=year"`
	Name       string `in:"query=name"`
	Surname    string `in:"query=surname"`
	Patronymic string `in:"query=patronymic"`
	Limit      int    `in:"query=limit"`
	Offset     int    `in:"query=offset"`
}