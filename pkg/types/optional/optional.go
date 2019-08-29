/*
Package optional is container for optional types.

Used to supply optional configuration more elegantly.
*/
package optional

//String is an optional string.
func String(str string) *string {
	return &str
}

//Int is an optional int.
func Int(i int) *int {
	return &i
}
