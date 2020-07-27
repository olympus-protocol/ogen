//+build dos_test

package dos_test

import "testing"

// DoS Test
// 1. A node is started without any block on the database but with a hardcoded genesis time.
// 2. It will try to proces multiple primitives that should be rejected and some other should pass.
func TestMain(m *testing.M) {

}