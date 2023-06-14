package httputil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCookie(t *testing.T) {
	const cookie1 = `exp=89cd78c2; enterprise_domain=wepack; MICRO_VERSION_SILENT=true; united=fb753123-5beb-48d6-b63c-789352a63ae8; cf=12cea2ca51c8557e81ec9de1088bd512; clientId=b94d737b-a79c-491e-bd96-580cea9bd831; SOME-TOKEN=dc7d7868-f07b-40fc-8dfb-7cc8fc3330ba; login=e866e424-d636-4e22-953c-cc368b77d16a; xid=1b0f8328f8774092bc5e78ef796332a7`
	cookies := ParseCookie(cookie1)
	require.Len(t, cookies, 9)

	for _, c := range cookies {
		if c.Name == "SOME-TOKEN" {
			assert.Equal(t, c.Value, "dc7d7868-f07b-40fc-8dfb-7cc8fc3330ba")
		}
	}
}
