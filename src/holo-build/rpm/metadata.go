/*******************************************************************************
*
* Copyright 2016 Stefan Majewsky <majewsky@gmx.net>
*
* This file is part of Holo.
*
* Holo is free software: you can redistribute it and/or modify it under the
* terms of the GNU General Public License as published by the Free Software
* Foundation, either version 3 of the License, or (at your option) any later
* version.
*
* Holo is distributed in the hope that it will be useful, but WITHOUT ANY
* WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
* A PARTICULAR PURPOSE. See the GNU General Public License for more details.
*
* You should have received a copy of the GNU General Public License along with
* Holo. If not, see <http://www.gnu.org/licenses/>.
*
*******************************************************************************/

package rpm

import (
	"fmt"
	"strings"

	"../common"
)

//MakeHeaderSection produces the header section of an RPM header.
func MakeHeaderSection(pkg *common.Package, payload *Payload) []byte {
	h := &Header{}

	addPackageInformationTags(h, pkg)
	h.AddInt32Value(RpmtagArchiveSize, []int32{int32(payload.UncompressedSize)})

	addInstallationTags(h, pkg)

	//TODO: [LSB, 25.2.4.3] file information tags
	//TODO: [LSB, 25.2.4.4] dependency information tags

	return h.ToBinary(RpmtagHeaderImmutable)
}

//see [LSB,25.2.4.1]
func addPackageInformationTags(h *Header, pkg *common.Package) {
	h.AddStringValue(RpmtagName, pkg.Name, false)
	h.AddStringValue(RpmtagVersion, versionString(pkg), false)
	h.AddStringValue(RpmtagRelease, fmt.Sprintf("%d", pkg.Release), false)

	//summary == first line of description
	descSplit := strings.SplitN(pkg.Description, "\n", 2)
	h.AddStringValue(RpmtagSummary, descSplit[0], true)
	h.AddStringValue(RpmtagDescription, pkg.Description, true)
	sizeInBytes := int32(pkg.FSRoot.InstalledSizeInBytes())
	h.AddInt32Value(RpmtagSize, []int32{sizeInBytes})

	//TODO validate that RPM implementations actually like this; there seems to
	//be no reference for how to spell "no license" in RPM (to compare, the
	//License attribute is optional in dpkg, and pacman accepts "custom:none")
	h.AddStringValue(RpmtagLicense, "None", false)

	if pkg.Author != "" {
		h.AddStringValue(RpmtagPackager, pkg.Author, false)
	}

	//source for valid package groups:
	//  <https://en.opensuse.org/openSUSE:Package_group_guidelines>
	//There is no such link for Fedora. Fedora treats the Group tag as optional
	//even though [LSB] says it's required. Source:
	//  <https://fedoraproject.org/wiki/Packaging:Guidelines?rd=Packaging/Guidelines#Group_tag>
	h.AddStringValue(RpmtagGroup, "System/Management", true)

	h.AddStringValue(RpmtagOs, "linux", false)
	h.AddStringValue(RpmtagArch, "noarch", false)

	h.AddStringValue(RpmtagPayloadFormat, "cpio", false)
	h.AddStringValue(RpmtagPayloadCompressor, "lzma", false)
	h.AddStringValue(RpmtagPayloadFlags, "5", false)
}

//see [LSB,25.2.4.2]
func addInstallationTags(h *Header, pkg *common.Package) {
	if pkg.SetupScript != "" {
		h.AddStringValue(RpmtagPostIn, pkg.SetupScript, false)
		h.AddStringValue(RpmtagPostInProg, "/bin/sh", false)
	}
	if pkg.CleanupScript != "" {
		h.AddStringValue(RpmtagPostUn, pkg.CleanupScript, false)
		h.AddStringValue(RpmtagPostUnProg, "/bin/sh", false)
	}
}
