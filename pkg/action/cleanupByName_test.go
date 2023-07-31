package action

import (
	"testing"
	"time"

	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/stretchr/testify/suite"
)

type ImageFilterSuite struct {
	suite.Suite
	cleanup    *CleanupByName
	commitList []string
}

func TestFillLists(t *testing.T) {
	suite.Run(t, &ImageFilterSuite{})
}

func (ifs *ImageFilterSuite) SetupTest() {
	ifs.cleanup = &CleanupByName{
		savedImages:       make(map[string]images.Image, 0),
		imagesForDeletion: make(map[string]images.Image, 0),
		scanDepth:         10,
		dryRun:            false,
	}
	ifs.commitList = []string{
		"ad6fed9464ef6f47b2d89ab856090d25c898d259",
		"f8b453a8b9dd6fd431577a47ec48f4ecf1500689",
		"26190eb145dca8fbb23bcd9967456a211545d459",
		"f67e14ac89404e564e856bc5c0e5f9f3c5608e16",
		"24a2165c150973fd87bc52633ecc2e6168a010b6",
		"5be8b85f8a27040845190b4e9ee5a7dd06222ab5",
		"9cc699a355a76df530a34be856d1418e52b3e6cd",
	}
}

func (ifs *ImageFilterSuite) TestFilterOK() {
	images := []images.Image{{
		ID:         "b9551daf-10df-4739-82a0-b7efc687e9c6",
		Tags:       []string{ifs.commitList[0], "master"},
		Visibility: "private",
		CreatedAt:  time.Now(),
	}, {
		ID:         "a66e2ab7-3de5-4cf3-bd24-104ccb511c8c",
		Tags:       []string{ifs.commitList[1], "master"},
		Visibility: "private",
		CreatedAt:  time.Now().Add(-time.Hour * 1),
	}, {
		ID:         "04f24cb4-beb0-4d87-b67a-d4834fba08ab",
		Tags:       []string{ifs.commitList[2], "master"},
		Visibility: "private",
		CreatedAt:  time.Now().Add(-time.Hour * 3),
	}}
	err := ifs.cleanup.filterImagesByCommitAndTime(images, ifs.commitList)

	ifs.Assert().Nil(err)
	ifs.Assert().Contains(ifs.cleanup.savedImages, images[0].ID)
	ifs.Assert().NotContains(ifs.cleanup.imagesForDeletion, images[0].ID)

	ifs.Assert().Contains(ifs.cleanup.imagesForDeletion, images[1].ID)
	ifs.Assert().Contains(ifs.cleanup.imagesForDeletion, images[2].ID)
}

func (ifs *ImageFilterSuite) TestFilterPublic() {
	images := []images.Image{{
		ID:         "e6637019-e80c-49b1-84ff-1bbe97cfcd64",
		Tags:       []string{ifs.commitList[0], "master"},
		Visibility: "private",
		CreatedAt:  time.Now(),
	}, {
		ID:         "5beb9780-8eed-480f-807f-7a99c89174f2",
		Tags:       []string{ifs.commitList[1], "master"},
		Visibility: "public",
		CreatedAt:  time.Now().Add(-time.Hour * 1),
	}, {
		ID:         "cf03fca9-e36b-4494-b8df-694d4cc4d319",
		Tags:       []string{},
		Visibility: "public",
		CreatedAt:  time.Now().Add(-time.Hour * 3),
	}}
	err := ifs.cleanup.filterImagesByCommitAndTime(images, ifs.commitList)

	ifs.Assert().Nil(err)

	for i := 0; i < len(images); i++ {
		ifs.Assert().Contains(ifs.cleanup.savedImages, images[i].ID)
		ifs.Assert().NotContains(ifs.cleanup.imagesForDeletion, images[i].ID)
	}
}

func (ifs *ImageFilterSuite) TestFilterByCommitInSameTime() {
	sameTime := time.Now()
	images := []images.Image{{
		ID:         "597c8284-d77f-4296-8f96-74028661ed81",
		Tags:       []string{ifs.commitList[0], "master"},
		Visibility: "private",
		CreatedAt:  time.Now().Add(-time.Hour * 3),
	}, {
		ID:         "bef78610-328a-4bd4-9a15-2c96d97ac933",
		Tags:       []string{ifs.commitList[0], "master"},
		Visibility: "private",
		CreatedAt:  sameTime,
	}, {
		ID:         "5427dbb9-4559-4abb-9496-0208890f61ce",
		Tags:       []string{ifs.commitList[1], "master"},
		Visibility: "private",
		CreatedAt:  sameTime,
	}}
	err := ifs.cleanup.filterImagesByCommitAndTime(images, ifs.commitList)

	ifs.Assert().Nil(err)

	ifs.Assert().Contains(ifs.cleanup.savedImages, images[1].ID)
	ifs.Assert().Contains(ifs.cleanup.imagesForDeletion, images[0].ID)
	ifs.Assert().Contains(ifs.cleanup.imagesForDeletion, images[2].ID)
	ifs.Assert().NotContains(ifs.cleanup.imagesForDeletion, images[1].ID)
}
