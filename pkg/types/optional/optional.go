package optional

//String optional string
func String(str string) *string {
	return &str
}

//Int optional int
func Int(i int) *int {
	return &i
}
