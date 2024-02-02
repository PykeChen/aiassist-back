package wxci

import "testing"

func TestImageUrl(t *testing.T) {
	url := "https://cfile.lukavoice.com/audit_freeze_backup/increment_audit/img/9BC5DBCB8605BFF1F87A64DCD97CC0FE.jpg"
	t.Log(wxciImageVeify(url))
}
