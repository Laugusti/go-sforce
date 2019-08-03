package credentials

import "fmt"

func ExampleNew() {
	creds := New("user@example.com", "P@s$w0rD", "test1n3289s32", `testjfa9a"afa8"'132%$#@@`)

	fmt.Println(creds)

	// Output:
	// &{user@example.com P@s$w0rD test1n3289s32 testjfa9a"afa8"'132%$#@@}
}
