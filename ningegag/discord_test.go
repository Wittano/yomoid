package ningegag

import "testing"

func TestFixNinegagFixingLink(t *testing.T) {
	data := map[string]string{
		"": "",
		"https://img-9gag-fun.9cache.com/photo/aMVvbeX_460sv.mp4":                                "https://img-9gag-fun.9cache.com/photo/aMVvbeX_460sv.mp4",
		"https://img-9gag-fun.9cache.com/photo/aGyG196_460svav1.mp4":                             "https://img-9gag-fun.9cache.com/photo/aGyG196_460sv.mp4",
		"https://media.discordapp.net/attachments/305375610052673536/1371507245493387285/2Q.png": "https://media.discordapp.net/attachments/305375610052673536/1371507245493387285/2Q.png",
	}

	for in, exp := range data {
		t.Run("Fix 9gag link: "+in, func(t *testing.T) {
			got, _ := fixNinegagLink(in)
			if exp != got {
				t.Fatalf("invalid 9gag link result. Expected: %s, got: %s", exp, got)
			}
		})
	}
}
