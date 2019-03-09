package hi

type HiService interface {
	Hi() (string)
}

type hiService struct{}

func (hiService) Hi() (string) {
	return "hello"
}

func NewService() HiService {
	return &hiService{}
}