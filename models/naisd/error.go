package naisd

type AppError interface {
	error
	Code() int
}

