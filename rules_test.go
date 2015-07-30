package adblock_test

import (
	"github.com/bluele/adblock"
	"strings"
	"testing"
)

func TestRule(t *testing.T) {
	rule, err := adblock.NewRule("||criteo.com^$third-party")
	if err != nil {
		t.Error("cannot parse")
	}
	if rule.MatchURL("http://cri.com", map[string]interface{}{
		"third-party": true,
	}) {
		t.Error("not match")
	}
}

func TestRules(t *testing.T) {
	lines := parseLines(resource)
	rules, err := adblock.NewRules(lines, nil)
	if err != nil {
		t.Error(err)
	}

	if !rules.ShouldBlock("http://jamo.tv/js/app.js", nil) {
		t.Error("should block")
	}
	if rules.ShouldBlock("http://google.com/ads/app.js", map[string]interface{}{
		"domain": "valid.example.com",
	}) {
		t.Error("should not block")
	}
	if !rules.ShouldBlock("http://google.com/ads/app.js", map[string]interface{}{
		"domain": "invalid.example.com",
	}) {
		t.Error("should block")
	}
}

func parseLines(text string) (lines []string) {
	for _, line := range strings.Split(text, "\n") {
		line = strings.Trim(line, "\r\n ")
		if len(line) != 0 {
			lines = append(lines, line)
		}
	}
	return
}

// https://easylist.adblockplus.org/en/
var resource = `
[Adblock Plus 2.0]
! Checksum: eQVrgYVjRUGJWOyE1JwG+A
! Version: 201507270530
! Title: EasyList
! Last modified: 27 Jul 2015 05:30 UTC
! Expires: 4 days (update frequency)
! Homepage: https://easylist.adblockplus.org/
! Licence: https://easylist-downloads.adblockplus.org/COPYING
!
! Please report any unblocked adverts or problems
! in the forums (https://forums.lanik.us/)
! or via e-mail (easylist.subscription@gmail.com).
!
!-----------------------General advert blocking filters-----------------------!
! *** easylist:easylist/easylist_general_block.txt ***
&ad_box_
&ad_channel=
@@||adultadworld.com/adhandler/$subdocument
@@||desihoes.com/advertisement.js
@@||fapxl.com^$elemhide
@@||fuqer.com^*/advertisement.js
@@||gaybeeg.info/wp-content/plugins/blockalyzer-adblock-counter/$image,domain=gaybeeg.info
@@||hentaienespañol.net^$elemhide
@@||hentaimoe.com/js/advertisement.js
@@||imgadult.com/js/advertisement.js
@@||indiangilma.com^$elemhide
@@||jav4.me^$script,domain=jav4.me
@@||javfee.com^$script,domain=javfee.com
@@||javpee.com/eroex.js
@@||jkhentai.tv^$script,domain=jkhentai.tv
@@||jporn4u.com/js/ads.js
@@||lfporn.com^$elemhide
@@||mongoporn.com^*/adframe/$subdocument
||jamo.tv
||google.com/ads/$domain=invalid.example.com`
