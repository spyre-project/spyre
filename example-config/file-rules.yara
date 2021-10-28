include "common.yara"

private rule PEFILE {
	condition:
		uint16(0) == 0x5a4d and uint32(uint32(0x3c)) == 0x4550
}

rule Packer_UPX {
	meta:
		description = "File packed with UPX"
		spyre_collect_limit = 0 // don't collect; -1=collect entire file; n>0=collect n bytes
	strings:
		$upx1 = {55505830000000}
		$upx2 = {55505831000000}
		$upx_sig = "UPX!"
	condition:
		PEFILE and
		for all of them : ( $ in (0..0x400) )
}

rule Archive_RAR_cloaked {
	meta:
		description = "RAR file cloaked by a different extension"
		license = "Detection Rule License 1.1 https://github.com/Neo23x0/signature-base/blob/master/LICENSE"
		author = "Florian Roth"
		spyre_collect_limit = 0xA00000
	condition:
		uint32be(0) == 0x52617221                           // RAR File Magic Header
		and not filename matches /(rarnew.dat|\.rar)$/is    // not the .RAR extension
		and not filename matches /\.[rR][\d]{2}$/           // split RAR file
		and not filepath contains "Recycle"                 // not a deleted RAR file in recycler
}

rule Archive_RAR_split {
	meta:
		spyre_collect_limit = 0xA00000
	condition:
		uint32be(0) == 0x52617221
		and (
			filename matches /\.part\d\d\.rar$/i
			or filename matches /\.r\d\d$/i
		)
}
