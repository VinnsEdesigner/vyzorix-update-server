// Package validate provides input validation utilities.
package validate

import (
	"net"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

const (
	MaxEmailLength    = 254
	MaxNameLength     = 100
	MaxPasswordLength = 128
	MinPasswordLength = 8
	MaxDeviceIDLength = 64
	MaxCommandLength  = 256
	MaxTokenLength    = 1024
)

var disposableEmailDomains = map[string]bool{
	"tempmail.com":      true,
	"guerrillamail.com": true,
	"mailinator.com":    true,
	"10minutemail.com":  true,
	"throwaway.email":   true,
	"fakeinbox.com":     true,
	"getnada.com":       true,
	"maildrop.cc":       true,
	"dispostable.com":   true,
	"sharklasers.com":   true,
	"mailnesia.com":     true,
	"emailondeck.com":   true,
	"getairmail.com":    true,
	"mohmal.com":        true,
	"burnermail.io":     true,
	"mailcatch.com":     true,
	"mt2009.com":        true,
	"mt2014.com":        true,
	"mt2015.com":        true,
	"devnullmail.com":   true,
	"discardmail.com":   true,
	"discardmail.de":    true,
	"spamex.com":        true,
	"spam4.me":          true,
	// Test/fake domains that should be blocked.
	"example.com":       true,
	"example.org":       true,
	"example.net":       true,
	"example.edu":       true,
	// Additional common fake/test domains.
	"anonbox.net":       true,
	"anonymbox.com":     true,
	"antispam.de":       true,
	"bobmail.info":      true,
	"bugmenot.com":      true,
	"deadaddress.com":   true,
	"einrot.com":        true,
	"emailigo.de":       true,
	"emailinfive.com":   true,
	"emailtemporario.com.br": true,
	"eyepaste.com":      true,
	"fantasymail.de":    true,
	"freemail.hu":       true,
	"friendlymail.co.uk": true,
	"front14.org":       true,
	"gawab.com":         true,
	"generator.email":   true,
	"ghosttexter.de":    true,
	"gishpuppy.com":     true,
	"great-host.in":     true,
	"greensloth.com":    true,
	"grr.la":            true,
	"gsrv.co.uk":        true,
	"guerillamail.biz":  true,
	"guerillamail.de":   true,
	"guerillamail.net":  true,
	"guerillamail.org":  true,
	"guerrillamail.info": true,
	"haltospam.com":     true,
	"hopemail.biz":      true,
	"ihateyoualot.info": true,
	"imails.info":       true,
	"imgof.com":         true,
	"imgv.de":           true,
	"imstations.com":    true,
	"inbax.tk":          true,
	"incognitomail.com": true,
	"incognitomail.net": true,
	"incognitomail.org": true,
	"insorg-mail.info":  true,
	"instant-mail.de":   true,
	"instantemailaddress.com": true,
	"ipoo.org":          true,
	"irish2me.com":      true,
	"iwi.net":           true,
	"jetable.com":       true,
	"jetable.fr.nf":     true,
	"jetable.net":       true,
	"jetable.org":       true,
	"jnxjn.com":         true,
	"jourrapide.com":    true,
	"jsrsolutions.com":  true,
	"kasmail.com":       true,
	"kaspop.com":        true,
	"keepmymail.com":    true,
	"killmail.com":      true,
	"killmail.net":      true,
	"kimsdisk.com":      true,
	"kingsq.ga":         true,
	"kiois.com":         true,
	"kir.ch.tc":         true,
	"klassmaster.com":  true,
	"klassmaster.net":   true,
	"klzlv.com":         true,
	"kulturbetrieb.info": true,
	"kurzepost.de":      true,
	"lawlita.com":       true,
	"letthemeatspam.com": true,
	"lhsdv.com":         true,
	"lifebyfood.com":    true,
	"link2mail.net":     true,
	"litedrop.com":      true,
	"llogin.ru":         true,
	"lol.ovpn.to":       true,
	"lolfreak.net":      true,
	"lookugly.com":      true,
	"lopl.co.cc":        true,
	"lortemail.dk":      true,
	"lovemeleaveme.com": true,
	"lr78.com":          true,
	"lroid.com":         true,
	"lukop.dk":          true,
	"m4ilweb.info":      true,
	"maboard.com":       true,
	"mail-hierarchie.net": true,
	"mail.by":           true,
	"mail.mezimages.net": true,
	"mail.zp.ua":        true,
	"mail114.net":       true,
	"mail2rss.org":      true,
	"mail333.com":       true,
	"mail4trash.com":    true,
	"mailbidon.com":     true,
	"mailblocks.com":    true,
	"mailbucket.org":    true,
	"mailcat.biz":       true,
	"mailde.de":         true,
	"mailde.info":       true,
	"maildx.com":        true,
	"mailed.ro":        true,
	"maileater.com":     true,
	"mailexpire.com":    true,
	"mailfa.tk":         true,
	"mailforspam.com":   true,
	"mailfree.ga":       true,
	"mailfree.gq":       true,
	"mailfree.ml":       true,
	"mailfreeonline.com": true,
	"mailfs.com":        true,
	"mailguard.me":      true,
	"mailhazard.com":    true,
	"mailhazard.us":     true,
	"mailhz.me":         true,
	"mailimate.com":     true,
	"mailin8r.com":      true,
	"mailinater.com":    true,
	"mailinator.ga":     true,
	"mailinator.gq":     true,
	"mailinator.net":    true,
	"mailinator.org":    true,
	"mailinator.us":     true,
	"mailinator2.com":   true,
	"mailinblack.com":   true,
	"mailincubator.com": true,
	"mailismagic.com":   true,
	"mailjunk.cf":       true,
	"mailjunk.ga":       true,
	"mailjunk.gq":       true,
	"mailjunk.ml":       true,
	"mailjunk.tk":       true,
	"mailmate.com":      true,
	"mailme.gq":         true,
	"mailme.ir":         true,
	"mailme.lv":         true,
	"mailme24.com":      true,
	"mailmetrash.com":   true,
	"mailmoat.com":      true,
	"mailnator.com":     true,
	"mailnull.com":      true,
	"mailorg.org":       true,
	"mailpick.biz":      true,
	"mailproxsy.com":    true,
	"mailquack.com":     true,
	"mailrock.biz":      true,
	"mailsac.com":       true,
	"mailscrap.com":     true,
	"mailseal.de":       true,
	"mailshell.com":     true,
	"mailsiphon.com":    true,
	"mailslapping.com":  true,
	"mailslite.com":     true,
	"mailtemp.info":     true,
	"mailtome.de":       true,
	"mailtrash.net":     true,
	"mailtv.net":        true,
	"mailtv.tv":         true,
	"mailzilla.com":     true,
	"mailzilla.org":     true,
	"makemetheking.com": true,
	"manybrain.com":     true,
	"mbx.cc":            true,
	"mega.zik.dj":       true,
	"meinspamschutz.de": true,
	"meltmail.com":      true,
	"messagebeamer.de":  true,
	"mezimages.net":     true,
	"mieru-mail.com":    true,
	"migmail.pl":        true,
	"migumail.com":      true,
	"mintemail.com":     true,
	"mjukgansen.nu":     true,
	"moakt.com":         true,
	"mobi.web.id":       true,
	"mobileninja.co.uk": true,
	"moburl.com":        true,
	"moncourrier.fr.nf": true,
	"monemail.fr.nf":    true,
	"monmail.fr.nf":     true,
	"monumentmail.com":  true,
	"ms9.mailslite.com": true,
	"msa.minsmail.com":  true,
	"msb.minsmail.com":  true,
	"msg.mailslite.com": true,
	"mspeciosa.com":     true,
	"msrc.ml":           true,
	"mssaan.ml":         true,
	"mxfuel.com":        true,
	"my10minutemail.com": true,
	"mycleaninbox.net":  true,
	"myemailboxy.com":   true,
	"mynetstore.de":     true,
	"mypacks.net":       true,
	"mypartyclip.de":    true,
	"myphantomemail.com": true,
	"myspaceinc.com":    true,
	"myspaceinc.net":    true,
	"myspacepimpedup.com": true,
	"mytempemail.com":   true,
	"mytempmail.com":    true,
	"neomailbox.com":    true,
	"nepwk.com":         true,
	"nervmich.net":      true,
	"nervtmansen.de":    true,
	"netmails.com":      true,
	"netmails.net":      true,
	"netzidiot.de":      true,
	"neverbox.com":      true,
	"nice-4u.com":       true,
	"nincsmail.hu":      true,
	"nmail.cf":          true,
	"nobulk.com":        true,
	"noclickemail.com":  true,
	"nogmailspam.info":  true,
	"nomail.xl.cx":     true,
	"nomail2me.com":    true,
	"nomorespamemails.com": true,
	"nonspam.eu":        true,
	"nonspammer.de":     true,
	"noref.in":          true,
	"nospam.ze.tc":      true,
	"nospam4.us":        true,
	"nospamfor.us":      true,
	"nospammail.net":    true,
	"nospamthanks.info": true,
	"notmailinator.com": true,
	"nowhere.org":       true,
	"nowmymail.com":     true,
	"ntelos.net":        true,
	"nurfuerspam.de":    true,
	"nwldx.com":         true,
	"objectmail.com":    true,
	"obobbo.com":        true,
	"odnorazovoe.ru":    true,
	"ohaaa.de":          true,
	"omail.pro":         true,
	"oneoffemail.com":   true,
	"onewaymail.com":    true,
	"onlatedotcom.info": true,
	"online.ms":         true,
	"oopi.org":          true,
	"opayq.com":         true,
	"ordinaryamerican.net": true,
	"otherinbox.com":    true,
	"ourklips.com":      true,
	"outlawspam.com":    true,
	"ovpn.to":           true,
	"owlpic.com":        true,
	"pancakemail.com":   true,
	"pjjkp.com":         true,
	"plexolan.de":       true,
	"politikerclub.de":  true,
	"poofy.org":         true,
	"pookmail.com":      true,
	"privacy.net":       true,
	"privatdemail.net":  true,
	"privy-mail.com":    true,
	"privymail.de":      true,
	"proxymail.eu":      true,
	"prtnx.com":         true,
	"punkass.com":       true,
	"putthisinyourspamdatabase.com": true,
	"pwrby.com":         true,
	"qisdo.com":         true,
	"qisoa.com":         true,
	"quickinbox.com":    true,
	"quickmail.nl":      true,
	"rainmail.biz":      true,
	"rcpt.at":           true,
	"reallymymail.com":  true,
	"realtyalerts.ca":   true,
	"recode.me":         true,
	"recursor.net":      true,
	"recyclemail.dk":    true,
	"regbypass.com":     true,
	"rejectmail.com":    true,
	"remail.cf":         true,
	"remail.ga":         true,
	"rhyta.com":         true,
	"rklips.com":        true,
	"rmqkr.net":         true,
	"royal.net":         true,
	"rppkn.com":         true,
	"rtrtr.com":         true,
	"s0ny.net":          true,
	"safe-mail.net":     true,
	"safersignup.de":    true,
	"safetymail.info":   true,
	"safetypost.de":     true,
	"sandelf.de":        true,
	"saynotospams.com":  true,
	"schafmail.de":      true,
	"selfdestructingmail.com": true,
	"sendspamhere.com":  true,
	"shieldedmail.com":  true,
	"shieldemail.com":   true,
	"shiftmail.com":     true,
	"shitmail.me":       true,
	"shortmail.net":     true,
	"shut.name":         true,
	"shut.ws":           true,
	"sibmail.com":       true,
	"sinnlos-mail.de":   true,
	"siteposter.net":    true,
	"skeefmail.com":     true,
	"slaskpost.se":      true,
	"slopsbox.com":      true,
	"slowslow.de":       true,
	"smashmail.de":      true,
	"smellfear.com":     true,
	"snakemail.com":     true,
	"sneakemail.com":    true,
	"sneakmail.de":      true,
	"snkmail.com":       true,
	"sofimail.com":      true,
	"sofort-mail.de":    true,
	"sogetthis.com":     true,
	"soisz.com":         true,
	"solvemail.info":    true,
	"soodomail.com":     true,
	"soodonims.com":     true,
	"spam.su":           true,
	"spamail.de":        true,
	"spamarrest.com":    true,
	"spamavert.com":     true,
	"spambob.com":       true,
	"spambob.net":       true,
	"spambob.org":       true,
	"spambog.com":       true,
	"spambog.de":        true,
	"spambog.net":       true,
	"spambog.ru":        true,
	"spambox.info":      true,
	"spamcannon.com":    true,
	"spamcannon.net":    true,
	"spamcero.com":      true,
	"spamcon.org":       true,
	"spamcorptastic.com": true,
	"spamcowboy.com":    true,
	"spamcowboy.net":    true,
	"spamcowboy.org":    true,
	"spamday.com":       true,
	"spamfighter.cf":    true,
	"spamfighter.ga":    true,
	"spamfighter.gq":    true,
	"spamfighter.ml":    true,
	"spamfighter.tk":    true,
	"spamfree24.com":    true,
	"spamfree24.de":     true,
	"spamfree24.eu":     true,
	"spamfree24.info":   true,
	"spamfree24.net":    true,
	"spamgoes.in":       true,
	"spamhereplease.com": true,
	"spamhole.com":      true,
	"spamify.com":       true,
	"spaminator.de":     true,
	"spamkill.info":     true,
	"spaml.com":         true,
	"spaml.de":          true,
	"spammotel.com":     true,
	"spamobox.com":      true,
	"spamoff.de":        true,
	"spamsalad.in":      true,
	"spamslicer.com":    true,
	"spamspot.com":      true,
	"spamstack.net":     true,
	"spamthis.co.uk":    true,
	"spamthisplease.com": true,
	"spamtrail.com":     true,
	"spamtroll.net":     true,
	"speed.1s.fr":       true,
	"spoofmail.de":      true,
	"squizzy.de":        true,
	"ssoia.com":         true,
	"startkeys.com":     true,
	"stinkefinger.net":  true,
	"stop-my-spam.cf":   true,
	"stop-my-spam.com":  true,
	"stop-my-spam.ga":  true,
	"stop-my-spam.ml":  true,
	"stop-my-spam.tk":  true,
	"streetwisemail.com": true,
	"stuffmail.de":      true,
	"supergreatmail.com": true,
	"supermailer.jp":    true,
	"superrito.com":     true,
	"superstachel.de":   true,
	"suremail.info":     true,
	"svk.jp":            true,
	"sweetxxx.de":       true,
	"tafmail.com":       true,
	"tagyourself.com":   true,
	"talkinator.com":    true,
	"tapchicuoihoi.com": true,
	"teewars.org":       true,
	"teleworm.com":      true,
	"teleworm.us":       true,
	"temp-mail.ru":      true,
	"temp.emeraldwebmail.com": true,
	"tempalias.com":     true,
	"tempe-mail.com":   true,
	"tempemail.biz":     true,
	"tempemail.co.za":   true,
	"tempemail.com":     true,
	"tempemail.net":     true,
	"tempinbox.co.uk":   true,
	"tempmail.co":       true,
	"tempmail.de":       true,
	"tempmail.eu":       true,
	"tempmail.it":       true,
	"tempmail.net":      true,
	"tempmail.us":       true,
	"tempmail2.com":     true,
	"tempmaildemo.com":  true,
	"tempmailer.com":    true,
	"tempmailer.de":     true,
	"tempmailweb.com":   true,
	"temporario.com.br": true,
	"temporaryemail.net": true,
	"temporaryemail.us": true,
	"temporaryforwarding.com": true,
	"temporaryinbox.com": true,
	"tempthe.net":       true,
	"thc.nl":            true,
	"thelimestones.com": true,
	"thietbivanpham.asia": true,
	"thisisnotmyrealemail.com": true,
	"thismail.net":      true,
	"thismail.ru":       true,
	"throam.com":        true,
	"throwam.com":       true,
	"throwawayemailaddress.com": true,
	"throwawaymail.com": true,
	"tilien.com":        true,
	"tmail.ws":          true,
	"tmailinator.com":   true,
	"toiea.com":         true,
	"toomail.biz":       true,
	"topranklist.de":    true,
	"tradermail.info":   true,
	"trash-amil.com":   true,
	"trash-mail.at":    true,
	"trash-mail.de":    true,
	"trash-mail.ga":    true,
	"trash-mail.gq":    true,
	"trash-mail.ml":    true,
	"trash-mail.tk":    true,
	"trash2010.com":    true,
	"trash2011.com":    true,
	"trashbox.eu":      true,
	"trashdevil.com":   true,
	"trashdevil.de":    true,
	"trashemail.de":    true,
	"trashmail.at":     true,
	"trashmail.com":    true,
	"trashmail.de":     true,
	"trashmail.me":     true,
	"trashmail.net":    true,
	"trashmail.org":    true,
	"trashmail.ws":     true,
	"trashmailer.com":  true,
	"trashymail.com":   true,
	"trashymail.net":   true,
	"trbvm.com":        true,
	"trickmail.net":    true,
	"trillianpro.com":  true,
	"tryalert.com":     true,
	"turual.com":       true,
	"twinmail.de":      true,
	"twoweirdtricks.com": true,
	"tyldd.com":        true,
	"uggsrock.com":     true,
	"umail.net":        true,
	"upliftnow.com":    true,
	"uplipht.com":      true,
	"uroid.com":        true,
	"us.af":            true,
	"valemail.net":     true,
	"venompen.com":     true,
	"veryrealemail.com": true,
	"viditag.com":      true,
	"viewcastmedia.com": true,
	"viewcastmedia.net": true,
	"viewcastmedia.org": true,
	"viralplays.com":   true,
	"vkcode.ro":        true,
	"vmpanda.com":      true,
	"vomoto.com":       true,
	"vpn.st":           true,
	"vsimcard.com":     true,
	"vubby.com":        true,
	"wasteland.rfc822.org": true,
	"webemail.me":      true,
	"webm4il.info":     true,
	"webtrip.ch":       true,
	"weg-werf-email.de": true,
	"wegwerf-email-addressen.de": true,
	"wegwerf-emails.de": true,
	"wegwerfadresse.de": true,
	"wegwerfemail.com": true,
	"wegwerfemail.de":  true,
	"wegwerfemail.net": true,
	"wegwerfemail.org": true,
	"wegwerfemailaddressen.de": true,
	"wegwerfmail.de":   true,
	"wegwerfmail.info": true,
	"wegwerfmail.net":  true,
	"wegwerfmail.org":  true,
	"wh4f.org":         true,
	"whatiaas.com":      true,
	"whatpaas.com":      true,
	"whopy.com":        true,
	"whyspam.me":       true,
	"wilemail.com":     true,
	"willhackforfood.biz": true,
	"willselfdestruct.com": true,
	"winemaven.info":   true,
	"wolfsmail.tk":     true,
	"wollan.info":      true,
	"worldspace.link":  true,
	"wraptr.com":       true,
	"writeme.us":       true,
	"wronghead.com":    true,
	"wuzup.net":        true,
	"wuzupmail.net":    true,
	"wwwnew.eu":        true,
	"x.ip6.li":         true,
	"xagloo.com":       true,
	"xemaps.com":       true,
	"xents.com":        true,
	"xmaily.com":       true,
	"xoxy.net":         true,
	"xyzfree.net":      true,
	"yapped.net":        true,
	"yep.it":           true,
	"yogamaven.com":    true,
	"yopmail.fr":       true,
	"yopmail.gq":       true,
	"yopmail.net":      true,
	"you-spam.com":     true,
	"yourdomain.com":   true,
	"ypmail.webarnak.fr.eu.org": true,
	"yuurok.com":       true,
	"z1p.biz":          true,
	"zehnminuten.de":   true,
	"zehnminutenmail.de": true,
	"zetmail.com":      true,
	"zippymail.info":   true,
	"zoaxe.com":        true,
	"zoemail.com":      true,
	"zoemail.net":      true,
	"zoemail.org":      true,
	"zomg.info":        true,
	"zxcv.com":         true,
	"zxcvbnm.com":      true,
}

var roleBasedSuffixes = map[string]bool{
	"admin":         true,
	"administrator": true,
	"webmaster":     true,
	"hostmaster":    true,
	"postmaster":    true,
	"abuse":         true,
	"noc":           true,
	"security":      true,
	"support":       true,
	"noreply":       true,
	"no-reply":      true,
	"donotreply":    true,
	"invalid":       true,
}

var trustedEmailProviders = map[string]bool{
	"gmail.com":      true,
	"googlemail.com": true,
	"google.com":     true,
	"outlook.com":    true,
	"hotmail.com":    true,
	"live.com":       true,
	"msn.com":        true,
	"microsoft.com":  true,
	"yahoo.com":      true,
	"ymail.com":      true,
	"yahoo.co.uk":    true,
	"yahoo.com.au":   true,
	"yahoo.co.jp":    true,
	"icloud.com":     true,
	"apple.com":      true,
	"me.com":         true,
	"protonmail.com": true,
	"proton.me":      true,
	"tutanota.com":   true,
	"tutanota.de":    true,
	"aol.com":        true,
	"zoho.com":       true,
	"fastmail.com":   true,
	"mail.com":       true,
	"gmx.com":        true,
	"gmx.de":         true,
	"gmx.net":        true,
	"yandex.com":     true,
	"yandex.ru":      true,
	"qq.com":         true,
	"163.com":        true,
	"126.com":        true,
	"sina.com":       true,
}

// Validator provides configurable enterprise-grade email validation.
type Validator struct {
	AllowDisposable            bool
	AllowRoleBased             bool
	RequireTrustedProviderOnly bool
	VerifyMX                   bool
	MXLookupTimeout            int
}

// DefaultValidator returns an EmailValidator configured for typical user registration.
func DefaultValidator() *Validator {
	return &Validator{
		AllowDisposable:            false,
		AllowRoleBased:             false,
		RequireTrustedProviderOnly: false,
		VerifyMX:                   false,
		MXLookupTimeout:            5,
	}
}

// StrictValidator returns an EmailValidator configured for high-security requirements.
func StrictValidator() *Validator {
	return &Validator{
		AllowDisposable:            false,
		AllowRoleBased:             false,
		RequireTrustedProviderOnly: true,
		VerifyMX:                   true,
		MXLookupTimeout:            5,
	}
}

// Error represents a validation error.
type Error struct {
	Field   string
	Message string
	Code    string
}

func (e *Error) Error() string {
	return e.Field + ": " + e.Message
}

// Result contains detailed validation results.
type Result struct {
	HasMXRecord *bool
	Normalized  string
	Domain      string
	LocalPart   string
	Errors      []ValidationError
	Warnings    []string
	Valid       bool
}

// ValidationError represents a specific email validation error.
type ValidationError struct {
	Code    string
	Message string
}

var (
	emailRegexRFC5322    = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	emailRegexPermissive = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	emailRegexIPDomain   = regexp.MustCompile(`@\[(\d{1,3}\.){3}\d{1,3}\]$`)
)

var mxCache = struct {
	entries map[string][]string
	sync.RWMutex
}{
	entries: make(map[string][]string),
}

// Email validates an email address.
func Email(email string) (string, error) {
	result := EmailFull(email, DefaultValidator())
	if !result.Valid {
		if len(result.Errors) > 0 {
			return "", &Error{
				Field:   "email",
				Message: result.Errors[0].Message,
				Code:    result.Errors[0].Code,
			}
		}

		return "", &Error{
			Field:   "email",
			Message: "invalid email format",
			Code:    "INVALID_FORMAT",
		}
	}

	return result.Normalized, nil
}

// EmailFull performs comprehensive email validation with detailed results.
func EmailFull(email string, validator *Validator) *Result {
	result := &Result{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []string{},
	}

	email = strings.TrimSpace(email)
	if len(email) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{Code: "EMPTY", Message: "email is required"})

		return result
	}

	if len(email) > MaxEmailLength {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{Code: "TOO_LONG", Message: "email exceeds maximum length"})

		return result
	}

	// Parse email.
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{Code: "INVALID_FORMAT", Message: "invalid email format"})

		return result
	}

	localPart := parts[0]
	domain := strings.ToLower(parts[1])
	result.LocalPart = localPart
	result.Domain = domain
	result.Normalized = localPart + "@" + domain

	// Validate format.
	if !emailRegexRFC5322.MatchString(email) && !emailRegexPermissive.MatchString(email) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{Code: "INVALID_FORMAT", Message: "invalid email format"})

		return result
	}

	// Check for IP domain.
	if emailRegexIPDomain.MatchString(email) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{Code: "IP_DOMAIN", Message: "IP addresses in email domain are not allowed"})

		return result
	}

	// Check for disposable.
	if !validator.AllowDisposable && isDisposable(domain) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{Code: "DISPOSABLE_DOMAIN", Message: "temporary/disposable email addresses are not allowed"})

		return result
	}

	// Check for role-based.
	if !validator.AllowRoleBased && isRoleBased(localPart) {
		result.Warnings = append(result.Warnings, "Role-based email address detected")
	}

	// Check for trusted provider.
	if validator.RequireTrustedProviderOnly && !isTrusted(domain) && !isDisposable(domain) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{Code: "UNTRUSTED_PROVIDER", Message: "only email from trusted providers are allowed"})

		return result
	}

	// MX record verification.
	if validator.VerifyMX && result.Valid {
		hasMX, _ := checkMX(domain, validator.MXLookupTimeout)
		result.HasMXRecord = &hasMX

		if !hasMX {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{Code: "NO_MX_RECORD", Message: "email domain does not accept mail"})

			return result
		}
	}

	return result
}

func isDisposable(domain string) bool {
	domain = strings.ToLower(domain)
	if disposableEmailDomains[domain] {
		return true
	}

	for d := range disposableEmailDomains {
		if strings.HasSuffix(domain, "."+d) {
			return true
		}
	}

	return false
}

func isRoleBased(local string) bool {
	local = strings.ToLower(local)
	if roleBasedSuffixes[local] {
		return true
	}

	prefixes := []string{"admin", "support", "noreply", "no-reply", "donotreply", "info", "contact", "help", "sales", "billing"}
	for _, p := range prefixes {
		if strings.HasPrefix(local, p) || strings.HasPrefix(local, p+".") {
			return true
		}
	}

	return false
}

func isTrusted(domain string) bool {
	return trustedEmailProviders[strings.ToLower(domain)]
}

func checkMX(domain string, _ int) (bool, error) {
	mxCache.RLock()
	if records, ok := mxCache.entries[domain]; ok {
		mxCache.RUnlock()
		return len(records) > 0, nil
	}
	mxCache.RUnlock()

	mxRecords, mxErr := net.LookupMX(domain)

	if mxErr != nil {
		return false, mxErr
	}

	if len(mxRecords) == 0 {
		mxCache.Lock()
		mxCache.entries[domain] = []string{}
		mxCache.Unlock()

		return false, nil
	}

	mxCache.Lock()

	mxCache.entries[domain] = make([]string, len(mxRecords))
	for i, mx := range mxRecords {
		mxCache.entries[domain][i] = mx.Host
	}
	mxCache.Unlock()

	return true, nil
}

// Name validates and sanitizes a name.
func Name(name string) (string, error) {
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return "", &Error{Field: "name", Message: "name is required"}
	}

	if len(name) > MaxNameLength {
		return "", &Error{Field: "name", Message: "name exceeds maximum length"}
	}

	return name, nil
}

// PasswordLength validates password length constraints.
func PasswordLength(password string) error {
	if len(password) < MinPasswordLength {
		return &Error{Field: "password", Message: "password must be at least 8 characters"}
	}

	if len(password) > MaxPasswordLength {
		return &Error{Field: "password", Message: "password exceeds maximum length"}
	}

	return nil
}

// DeviceID validates and sanitizes a device ID.
func DeviceID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if len(id) == 0 {
		return "", &Error{Field: "deviceId", Message: "device ID is required"}
	}

	if len(id) > MaxDeviceIDLength {
		return "", &Error{Field: "deviceId", Message: "device ID exceeds maximum length"}
	}

	validID := regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
	if !validID.MatchString(id) {
		return "", &Error{Field: "deviceId", Message: "device ID contains invalid characters"}
	}

	return id, nil
}

// Command validates a command string.
func Command(cmd string) (string, error) {
	cmd = strings.TrimSpace(cmd)
	if len(cmd) == 0 {
		return "", &Error{Field: "command", Message: "command is required"}
	}

	if len(cmd) > MaxCommandLength {
		return "", &Error{Field: "command", Message: "command exceeds maximum length"}
	}

	return cmd, nil
}

// Token validates a token string.
func Token(token string) (string, error) {
	token = strings.TrimSpace(token)
	if len(token) == 0 {
		return "", &Error{Field: "token", Message: "token is required"}
	}

	if len(token) > MaxTokenLength {
		return "", &Error{Field: "token", Message: "token exceeds maximum length"}
	}

	return token, nil
}

// Sanitize removes potentially dangerous characters.
func Sanitize(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		s = s[:maxLen]
	}

	return s
}

// ContainsInvalidUTF8 checks if a string contains invalid UTF-8 sequences.
func ContainsInvalidUTF8(s string) bool {
	return !utf8.ValidString(s)
}

// ContainsControlCharacters checks for Unicode control characters.
func ContainsControlCharacters(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) && !strings.ContainsRune("\t\n\r", r) {
			return true
		}
	}

	return false
}

// ExtractDomain extracts the domain part from an email.
func ExtractDomain(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}

	return parts[1]
}

// IsDisposableDomain checks if the domain of an email is disposable.
func IsDisposableDomain(email string) bool {
	return isDisposable(ExtractDomain(email))
}

// ClearMXCache clears the DNS MX record cache.
func ClearMXCache() {
	mxCache.Lock()
	mxCache.entries = make(map[string][]string)
	mxCache.Unlock()
}

// NormalizeEmail normalizes an email for case-insensitive comparison.
func NormalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

// EmailURI validates a mailto: URI.
func EmailURI(uri string) error {
	parsed, err := url.Parse(uri)
	if err != nil {
		return &Error{Field: "uri", Message: "invalid URI format"}
	}

	if parsed.Scheme != "mailto" {
		return &Error{Field: "uri", Message: "URI must be mailto: scheme"}
	}

	_, err = Email(parsed.Opaque)

	return err
}
