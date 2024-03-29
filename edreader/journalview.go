package edreader

import (
	"log"
	"sort"
	"strings"

	"github.com/vaegue/EDx52display_Redux/edsm"
	"github.com/vaegue/EDx52display_Redux/mfd"
)

// RefreshDisplay updates the display with the current state
func RefreshDisplay(state Journalstate) {
	MfdLock.Lock()
	defer MfdLock.Unlock()
	Mfd.Pages[pageLocation] = mfd.NewPage()
	renderLocationPage(&Mfd.Pages[pageLocation], state)
	Mfd.Pages[pageTargetInfo] = mfd.NewPage()
	renderFSDTarget(&Mfd.Pages[pageTargetInfo], state)
}

func renderLocationPage(page *mfd.Page, state Journalstate) {
	if state.Type == LocationPlanet || state.Type == LocationLanded {
		renderEDSMBody(page, "P", state.Location.Body, state.Location.SystemAddress, state.BodyID)
	} else {
		renderEDSMSystem(page, "S", state.Location.StarSystem, state.Location.SystemAddress)
	}

}

func renderFSDTarget(page *mfd.Page, state Journalstate) {
	if state.EDSMTarget.SystemAddress == 0 {
		page.Add("No FSD Target")
	} else {
		renderEDSMSystem(page, "T", state.EDSMTarget.Name, state.EDSMTarget.SystemAddress)
	}
}

func renderEDSMSystem(page *mfd.Page, header, systemname string, systemaddress int64) {
	sysinfopromise := edsm.GetSystemBodies(systemaddress)
	valueinfopromise := edsm.GetSystemValue(systemaddress)

	sysinfo := <-sysinfopromise

	sys := sysinfo.S

	if sysinfo.Error != nil {
		log.Println("Unable to fetch system information: ", sysinfo.Error)
		page.Add("Sysinfo lookup error")
		return
	}

	if sys.ID64 == 0 {
		page.Add("No EDSM data")
		return
	}

	mainBody := sys.MainStar()
	if mainBody.IsScoopable {
		page.Add("%s{%s", header, systemname)
	} else {
		page.Add("%s-%s", header, systemname)
	}

	// page.Add(mainBody.SubType)

	// page.Add("Bodies: %d", sys.BodyCount)

	valinfo := <-valueinfopromise
	if valinfo.Error == nil {
		page.Add(printer.Sprintf("S:%d M:%d", valinfo.S.EstimatedValue, valinfo.S.EstimatedValueMapped))
		// page.Add(printer.Sprintf("Map:%12d", valinfo.S.EstimatedValueMapped))

		if len(valinfo.S.ValuableBodies) > 0 {
			page.Add("Valuable Bodies:")
		}
		for _, valbody := range valinfo.S.ValuableBodies {
			bname := valbody.ShortName(sys)
			valstr := printer.Sprintf("%d", valbody.ValueMax)
			pad := 1
			if len(bname)+len(valstr) < 16 {
				pad = 16 - (len(bname) + len(valstr))
			}
			padstr := strings.Repeat(" ", pad)
			page.Add("%s%s%s", bname, padstr, valstr)
		}
	}

	landables := []edsm.Body{}
	matLocations := map[string][]edsm.Body{}

	for _, b := range sys.Bodies {
		if b.IsLandable {
			landables = append(landables, b)
			for m := range b.Materials {
				mlist, ok := matLocations[m]
				if !ok {
					mlist = []edsm.Body{}
					matLocations[m] = mlist
				}
				matLocations[m] = append(mlist, b)
			}
		}
	}

	if len(landables) == 0 {
		return
	}

	page.Add("Prospecting:")
	matlist := []string{}
	for mat := range matLocations {
		matlist = append(matlist, mat)
		bodies := matLocations[mat]
		sort.Slice(bodies, func(i, j int) bool { return bodies[i].Materials[mat] > bodies[j].Materials[mat] })
	}

	sort.Slice(matlist, func(i, j int) bool {
		matA := matlist[i]
		matB := matlist[j]
		a := matLocations[matA]
		b := matLocations[matB]
		if len(a) == len(b) {
			return a[0].Materials[matA] > b[0].Materials[matB]
		}
		return len(a) > len(b)

	})
	for _, mat := range matlist {
		bodies := matLocations[mat]
		page.Add("%s %d", mat, len(bodies))
		b := bodies[0]
		page.Add("%s: %.2f%%", b.ShortName(sys), b.Materials[mat])
	}

	return
}

func renderEDSMBody(page *mfd.Page, header, bodyName string, systemaddress, bodyid int64) {
	sysinfopromise := edsm.GetSystemBodies(systemaddress)
	page.Add(header)
	page.Add(bodyName)
	sysinfo := <-sysinfopromise
	if sysinfo.Error != nil {
		log.Println("Unable to fetch system information: ", sysinfo.Error)
		page.Add("Sysinfo lookup error")
		return
	}
	sys := sysinfo.S
	if sys.ID64 == 0 {
		page.Add("No EDSM data")
		return
	}

	body := sys.BodyByID(bodyid)
	if body.BodyID == 0 {
		page.Add("No EDSM data")
		return
	}

	page.Add("Gravity %7.2fG", body.Gravity)

	page.Add("Materials:")
	for _, m := range body.MaterialsSorted() {
		page.Add("%5.2f%% %s", m.Percentage, m.Name)
	}

	return
}
