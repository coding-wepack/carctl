package repo

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRepoProxySourceList(t *testing.T) {
	const proxyUrl1 = "http://wepack.cxy.dev.coding.io/api/user/wepack/project/registry/artifacts/repositories/3/proxies"
	const cookie1 = "exp=89cd78c2; code=artifact-reforge%3Dfalse%2Casync-blocked%3Dtrue%2Cauth-by-wechat%3Dtrue%2Cci-qci%3Dfalse%2Cci-team-step%3Dfalse%2Cci-team-templates%3Dfalse%2Ccoding-flow%3Dfalse%2Ccoding-ocd-java%3Dfalse%2Ccoding-ocd-pages%3Dtrue%2Centerprise-permission-management%3Dtrue%2Cmobile-layout-test%3Dfalse%2Cproject-permission-management%3Dtrue%2Cservice-exception-tips%3Dfalse%2Ctencent-cloud-object-storage%3Dtrue%2C5b585a51; enterprise_domain=wepack; MICRO_VERSION_SILENT=true; united=fb753b34-5beb-48d6-b63c-789352a63ae8; x_host_key_access=42f436806ec83550ba2dff4e7d06c37cda6201f9_s; x-client-ssid=188b2cb4d6d-9aab8a98d003b213bc4e5c762862408b506ad8fa; c=auth-by-wechat%3Dtrue%2Cproject-permission-management%3Dtrue%2Centerprise-permission-management%3Dtrue%2C5c58505d; cf=12cea2ca51c8557e81ec9de1088bd500; clientId=b94d737b-a79c-491e-bd96-580cea9bd834; XSRF-TOKEN=dc7d7868-f07b-40fc-8dfb-7cc8fc3330ba; login=e866e424-d636-4e22-953c-cc368b77d16a; coding_utm_source=Login%20Success; eid=1b0f8328-f877-4092-bc5e-78ef796332a6; coding_demo_visited=1"

	proxySources, err := GetRepoProxySourceList(proxyUrl1, cookie1)
	require.NoError(t, err)
	require.Len(t, proxySources, 3)
	fmt.Printf("%#v\n", proxySources)
}
