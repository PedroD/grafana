package sqlstore

import (
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/grafana/grafana/pkg/components/simplejson"
	m "github.com/grafana/grafana/pkg/models"
)

func updateTestDashboard(dashboard *m.Dashboard, data map[string]interface{}) {
	data["title"] = dashboard.Title

	saveCmd := m.SaveDashboardCommand{
		OrgId:     dashboard.OrgId,
		Overwrite: true,
		Dashboard: simplejson.NewFromAny(data),
	}

	err := SaveDashboard(&saveCmd)
	So(err, ShouldBeNil)
}

func TestGetDashboardVersion(t *testing.T) {
	Convey("Testing dashboard version retrieval", t, func() {
		InitTestDB(t)

		Convey("Get a Dashboard ID and version ID", func() {
			savedDash := insertTestDashboard("test dash 26", 1, "diff")

			query := m.GetDashboardVersionQuery{
				DashboardId: savedDash.Id,
				Version:     savedDash.Version,
			}

			err := GetDashboardVersion(&query)
			So(err, ShouldBeNil)
			So(savedDash.Id, ShouldEqual, query.DashboardId)
			So(savedDash.Version, ShouldEqual, query.Version)

			dashCmd := m.GetDashboardQuery{
				OrgId: savedDash.OrgId,
				Slug:  savedDash.Slug,
			}
			err = GetDashboard(&dashCmd)
			So(err, ShouldBeNil)
			eq := reflect.DeepEqual(dashCmd.Result.Data, query.Result.Data)
			So(eq, ShouldEqual, true)
		})

		Convey("Attempt to get a version that doesn't exist", func() {
			query := m.GetDashboardVersionQuery{
				DashboardId: int64(999),
				Version:     123,
			}

			err := GetDashboardVersion(&query)
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, m.ErrDashboardVersionNotFound)
		})
	})
}

func TestGetDashboardVersions(t *testing.T) {
	Convey("Testing dashboard versions retrieval", t, func() {
		InitTestDB(t)
		savedDash := insertTestDashboard("test dash 43", 1, "diff-all")

		Convey("Get all versions for a given Dashboard ID", func() {
			query := m.GetDashboardVersionsQuery{DashboardId: savedDash.Id, OrgId: 1}

			err := GetDashboardVersions(&query)
			So(err, ShouldBeNil)
			So(len(query.Result), ShouldEqual, 1)
		})

		Convey("Attempt to get the versions for a non-existent Dashboard ID", func() {
			query := m.GetDashboardVersionsQuery{DashboardId: int64(999), OrgId: 1}

			err := GetDashboardVersions(&query)
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, m.ErrNoVersionsForDashboardId)
			So(len(query.Result), ShouldEqual, 0)
		})

		Convey("Get all versions for an updated dashboard", func() {
			updateTestDashboard(savedDash, map[string]interface{}{
				"tags": "different-tag",
			})

			query := m.GetDashboardVersionsQuery{DashboardId: savedDash.Id, OrgId: 1}
			err := GetDashboardVersions(&query)

			So(err, ShouldBeNil)
			So(len(query.Result), ShouldEqual, 2)
		})
	})
}

func TestCompareDashboardVersions(t *testing.T) {
	Convey("Testing dashboard version comparison", t, func() {
		InitTestDB(t)

		savedDash := insertTestDashboard("test dash 43", 1, "x")
		updateTestDashboard(savedDash, map[string]interface{}{
			"tags": "y",
		})

		Convey("Compare two versions that are different", func() {
			query := m.GetDashboardVersionsQuery{DashboardId: savedDash.Id, OrgId: 1}

			err := GetDashboardVersions(&query)
			So(err, ShouldBeNil)
			So(len(query.Result), ShouldEqual, 2)

			cmd := m.CompareDashboardVersionsCommand{
				DashboardId: savedDash.Id,
				BaseVersion: query.Result[0].Version,
				NewVersion:  query.Result[1].Version,
				DiffType:    m.DiffDelta,
			}

			err = CompareDashboardVersionsCommand(&cmd)
			So(err, ShouldBeNil)
			So(cmd.Delta, ShouldNotBeNil)
		})

		Convey("Compare two versions that are the same", func() {
			cmd := m.CompareDashboardVersionsCommand{
				DashboardId: savedDash.Id,
				BaseVersion: savedDash.Version,
				NewVersion:  savedDash.Version,
				DiffType:    m.DiffDelta,
			}

			err := CompareDashboardVersionsCommand(&cmd)
			So(err, ShouldNotBeNil)
			So(cmd.Delta, ShouldBeNil)
		})

		Convey("Compare two versions that don't exist", func() {
			cmd := m.CompareDashboardVersionsCommand{
				DashboardId: savedDash.Id,
				BaseVersion: 123,
				NewVersion:  456,
				DiffType:    m.DiffDelta,
			}

			err := CompareDashboardVersionsCommand(&cmd)
			So(err, ShouldNotBeNil)
			So(cmd.Delta, ShouldBeNil)
		})
	})
}