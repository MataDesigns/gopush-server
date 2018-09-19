package storage

const (
	// TotalCountKey is key name for total count of storage
	TotalCountKey = "gorush-total-count"

	// IosSuccessKey is key name or ios success count of storage
	IosSuccessKey = "gorush-ios-success-count"

	// IosErrorKey is key name or ios success error of storage
	IosErrorKey = "gorush-ios-error-count"

	// AndroidSuccessKey is key name for android success count of storage
	AndroidSuccessKey = "gorush-android-success-count"

	// AndroidErrorKey is key name for android error count of storage
	AndroidErrorKey = "gorush-android-error-count"
)

type Storage interface {
	Reset()
	GetAllKeys(keys *[]string)
	GetAll(values *[]string)
	Get(string, *string) error
	GetInt(string, *int64)
	Set(string, string)
	SetInt(string, int64)
}
