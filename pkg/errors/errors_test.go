package errors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

//OpTests composes a testsuite to run all the Ops related tests in one group
type OpTests struct {
	suite.Suite
}

func TestErrMsg(t *testing.T) {
	const op Op = "TestErrMsg"
	msg := "test error"
	err := E(op, msg)
	require.Equal(t, msg, err.Error())
}

func TestErrEmbed(t *testing.T) {
	const op Op = "TestErrEmbed"
	msg := "test error"
	childErr := errors.New(msg)
	err := E(op, childErr)
	require.Equal(t, err.Error(), childErr.Error())
}

func TestSeverity(t *testing.T) {
	const op Op = "TestSeverity"
	msg := "test error"

	err := E(op, msg)
	require.Equal(t, logrus.ErrorLevel, Severity(err))

	err = E(op, msg, logrus.WarnLevel)
	require.Equal(t, logrus.WarnLevel, Severity(err))

	err = E(op, err)
	require.Equal(t, logrus.WarnLevel, Severity(err))

	err = E(op, err, logrus.InfoLevel)
	require.Equal(t, logrus.InfoLevel, Severity(err))
}

func TestKind(t *testing.T) {
	const op Op = "TestKind"
	msg := "test error"

	err := E(op, msg, KindBadRequest)
	require.Equal(t, KindBadRequest, Kind(err))
	require.Equal(t, http.StatusText(http.StatusBadRequest), KindText(err))
}

func TestOps(t *testing.T) {
	suite.Run(t, new(OpTests))
}

func (op *OpTests) TestOps() {
	const (
		op1 Op = "TestOps.op1"
		op2 Op = "TestOps.op2"
		op3 Op = "TestOps.op3"
	)

	err1 := E(op1, "op 1")
	err2 := E(op2, err1)
	err3 := E(op3, err2)

	require.ElementsMatch(op.T(), []Op{op1, op2, op3}, Ops(err3.(Error)))
}

func (op *OpTests) SetupTest() {}

func (op *OpTests) TestString() {
	const op1 Op = "testOps.op1"
	require.Equal(op.T(), op1.String(), "testOps.op1")
}
