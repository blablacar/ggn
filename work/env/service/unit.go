package service

import (
	"bufio"
	"github.com/Sirupsen/logrus"
	"github.com/blablacar/cnt/utils"
	"github.com/blablacar/ggn/spec"
	"github.com/coreos/fleet/unit"
	"github.com/juju/errors"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type Unit struct {
	Log      logrus.Entry
	path     string
	Name     string
	Filename string
	unitPath string
	Service  spec.Service
}

func NewUnit(path string, hostname string, service spec.Service) *Unit {
	l := service.GetLog()

	unitInfo := strings.Split(hostname, "_")
	if len(unitInfo) != 3 {
	}

	filename := service.GetEnv().GetName() + "_" + service.GetName() + "_" + hostname + ".service"
	unit := &Unit{
		Log:      *l.WithField("unit", hostname),
		Service:  service,
		path:     path,
		Name:     hostname,
		Filename: filename,
		unitPath: path + "/" + filename,
	}
	return unit
}

func (u Unit) Check() {
	u.Log.Debug("Check")

	u.Service.GetEnv().RunEarlyHook(u.Name, "check")
	defer u.Service.GetEnv().RunLateHook(u.Name, "check")

	statuses := u.Service.GetEnv().ListUnits()
	var status spec.UnitStatus
	if _, ok := statuses[u.Filename]; !ok {
		u.Log.Warn("cannot find unit on fleet")
		return
	}
	status = statuses[u.Filename]
	logrus.WithField("status", status).Debug("status")

	if status.Active != spec.ACTIVE_ACTIVE {
		u.Log.WithField("active", status.Active).Warn("unit status is not active")
		return
	}
	if status.Sub != spec.SUB_RUNNING {
		u.Log.WithField("sub", status.Sub).Warn("unit sub is not running")
		return
	}

	same, err := u.IsLocalContentSameAsRemote()
	if err != nil {
		u.Log.Error("Cannot read unit")
		return
	}
	if !same {
		u.Log.Warn("Unit is not up to date")
		return
	}
}

func (u Unit) Journal(follow bool, lines int) {
	u.Log.Debug("journal")

	args := []string{"journal", "-lines", strconv.Itoa(lines)}
	if follow {
		args = append(args, "-f")
	}
	args = append(args, u.Filename)

	err := u.Service.GetEnv().RunFleetCmd(args...)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to run journal")
	}
}

func (u Unit) Ssh() {
	u.Log.Debug("ssh")

	err := u.Service.GetEnv().RunFleetCmd("ssh", u.Filename)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to run status")
	}
}

func (u Unit) Diff() {
	u.Log.Debug("diff")

	same, err := u.IsLocalContentSameAsRemote()
	if err != nil {
		u.Log.Warn("Cannot read unit")
	}
	if !same {
		u.DisplayDiff()
	}
}

func (u Unit) GetUnitContentAsFleeted() (string, error) {
	unitFileContent, err := ioutil.ReadFile(u.unitPath)
	if err != nil {
		return "", err
	}

	fleetunit, err := unit.NewUnitFile(string(unitFileContent))
	if err != nil {
		return "", err
	}
	return convertMultilineUnitToString([]byte(fleetunit.String())), nil
}

func (u Unit) Status() error {
	u.Log.Debug("status")

	content, _, err := u.Service.GetEnv().RunFleetCmdGetOutput("status", u.Filename)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to run status")
	}
	os.Stdout.WriteString(content)
	return nil
	//	if err != nil {
	//		return "", err
	//	}
	//
	//	reg, err := regexp.Compile(`Active: (active|inactive|deactivating|activating)`)
	//	if err != nil {
	//		u.Log.Panic("Cannot compile regex")
	//	}
	//	matched := reg.FindStringSubmatch(content)
	//
	//	//	if !strings.Contains(content, "Active: %s ") { // Active: failed
	//	//		return content, errors.New("unit is not in running state")
	//	//	}
	//
	//	return matched[1], err
}

func (u Unit) DisplayDiff() error {
	u.Log.Debug("Diff")

	local, remote, err := u.serviceLocalAndRemoteContent()
	if err != nil {
		return err
	}

	localPath := "/tmp/" + u.Name + "__local"
	remotePath := "/tmp/" + u.Name + "__remote"

	ioutil.WriteFile(localPath, []byte(local), 0644)
	defer os.Remove(localPath)
	ioutil.WriteFile(remotePath, []byte(remote), 0644)
	defer os.Remove(remotePath)
	return utils.ExecCmd("git", "diff", "--word-diff", remotePath, localPath)
}

func (u Unit) IsLocalContentSameAsRemote() (bool, error) {
	local, remote, err := u.serviceLocalAndRemoteContent()
	if local != "" && err != nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return local == remote, nil
}

///////////////////////////////////////////

func (u Unit) serviceLocalAndRemoteContent() (string, string, error) {
	localContent, err := u.GetUnitContentAsFleeted()
	if err != nil {
		u.Log.WithError(err).Debug("Cannot get local unit content")
		return "", "", errors.Annotate(err, "Cannot read local unit file")
	}

	remoteContent, err := u.Service.GetFleetUnitContent(u.Filename)
	if err != nil {
		return localContent, "", errors.Annotate(err, "CanCannot read remote unit file")
	}
	return localContent, remoteContent, nil
}

func convertMultilineUnitToString(content []byte) string {
	var lines []string
	var currentLine string
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" && currentLine != "" {
			currentLine = strings.TrimRight(currentLine, " ")
		}
		currentLine += strings.TrimRight(line, " ")
		if strings.HasSuffix(currentLine, "\\") {
			currentLine = currentLine[0:len(currentLine)-2] + "  "
		} else {
			lines = append(lines, currentLine)
			currentLine = ""
		}
	}
	return strings.Join(lines, "\n")
}
