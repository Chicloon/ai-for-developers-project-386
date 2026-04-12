package api

import (
	"context"
	"net/http"
	"testing"

	"call-booking/internal/models"
	"call-booking/internal/uuid"
)

func TestGroupsList_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)

	// Create user
	email := "groups@example.com"
	userID := createTestUser(t, db, email, "password123", "Groups Test User")
	token := getAuthToken(userID, email)

	rr := makeRequest(router, "GET", "/api/my/groups", nil, token)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string][]models.VisibilityGroup
	parseResponse(t, rr, &resp)

	// Auto-created groups should exist (family, work, friends)
	if len(resp["groups"]) != 3 {
		t.Errorf("expected 3 auto-created groups, got %d", len(resp["groups"]))
	}
}

func TestGroupsList_WithGroups(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)

	email := "groups2@example.com"
	userID := createTestUser(t, db, email, "password123", "Groups Test User")
	token := getAuthToken(userID, email)

	rr := makeRequest(router, "GET", "/api/my/groups", nil, token)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string][]models.VisibilityGroup
	parseResponse(t, rr, &resp)

	if len(resp["groups"]) != 3 {
		t.Errorf("expected 3 groups from registration, got %d", len(resp["groups"]))
	}

	familyFound, workFound, friendsFound := false, false, false
	for _, g := range resp["groups"] {
		switch g.VisibilityLevel {
		case "family":
			familyFound = true
		case "work":
			workFound = true
		case "friends":
			friendsFound = true
		}
	}
	if !familyFound || !workFound || !friendsFound {
		t.Error("expected family, work, and friends groups")
	}
}

func TestGroupsList_Unauthorized(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	router := NewRouter(db)

	rr := makeRequest(router, "GET", "/api/my/groups", nil, "")

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestGroupsListMembers_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create user
	email := "groups11@example.com"
	userID := createTestUser(t, db, email, "password123", "Groups Test User")
	token := getAuthToken(userID, email)

	// Insert a group (using friends level since public was removed)
	var groupID string
	err := db.QueryRowContext(ctx, "INSERT INTO visibility_groups (id, owner_id, name, visibility_level) VALUES (?, ?, 'Empty Group', 'friends') RETURNING id", uuid.New(), userID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %v", err)
	}

	rr := makeRequest(router, "GET", "/api/my/groups/"+groupID+"/members", nil, token)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string][]models.GroupMember
	parseResponse(t, rr, &resp)

	if resp["members"] == nil {
		t.Error("expected empty array, not nil")
	}
	if len(resp["members"]) != 0 {
		t.Errorf("expected 0 members, got %d", len(resp["members"]))
	}
}

func TestGroupsListMembers_WithMembers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create owner user
	ownerEmail := "groupowner3@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Group Owner")
	ownerToken := getAuthToken(ownerID, ownerEmail)

	// Create member user
	memberEmail := "groupmember@example.com"
	memberID := createTestUser(t, db, memberEmail, "password123", "Group Member")

	// Insert a group
	var groupID string
	err := db.QueryRowContext(ctx, "INSERT INTO visibility_groups (id, owner_id, name, visibility_level) VALUES (?, ?, 'Group with Members', 'family') RETURNING id", uuid.New(), ownerID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %v", err)
	}

	// Add member to group
	_, err = db.ExecContext(ctx, "INSERT INTO group_members (id, group_id, member_id, added_by) VALUES (?, ?, ?, ?)", uuid.New(), groupID, memberID, ownerID)
	if err != nil {
		t.Fatalf("failed to add member: %v", err)
	}

	rr := makeRequest(router, "GET", "/api/my/groups/"+groupID+"/members", nil, ownerToken)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string][]models.GroupMember
	parseResponse(t, rr, &resp)

	if len(resp["members"]) != 1 {
		t.Errorf("expected 1 member, got %d", len(resp["members"]))
	}

	if len(resp["members"]) > 0 {
		member := resp["members"][0]
		if member.Member.ID != memberID {
			t.Errorf("expected member ID %s, got %s", memberID, member.Member.ID)
		}
		if member.Member.Email != memberEmail {
			t.Errorf("expected member email %s, got %s", memberEmail, member.Member.Email)
		}
	}
}

func TestGroupsListMembers_NotOwner(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create two users
	ownerEmail := "groupowner4@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Group Owner")

	otherEmail := "groupother3@example.com"
	otherID := createTestUser(t, db, otherEmail, "password123", "Other User")
	otherToken := getAuthToken(otherID, otherEmail)

	// Insert a group for owner
	var groupID string
	err := db.QueryRowContext(ctx, "INSERT INTO visibility_groups (id, owner_id, name, visibility_level) VALUES (?, ?, 'Private Group', 'family') RETURNING id", uuid.New(), ownerID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %v", err)
	}

	rr := makeRequest(router, "GET", "/api/my/groups/"+groupID+"/members", nil, otherToken)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rr.Code)
	}

	errResp := parseErrorResponse(t, rr)
	if errResp.Error != "you don't have access to this group" {
		t.Errorf("unexpected error message: %s", errResp.Error)
	}
}

func TestGroupsAddMember_ByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create owner user
	ownerEmail := "groupowner5@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Group Owner")
	ownerToken := getAuthToken(ownerID, ownerEmail)

	// Create user to add as member
	memberEmail := "membertoadd@example.com"
	memberID := createTestUser(t, db, memberEmail, "password123", "Member to Add")

	// Insert a group
	var groupID string
	err := db.QueryRowContext(ctx, "INSERT INTO visibility_groups (id, owner_id, name, visibility_level) VALUES (?, ?, 'Group', 'work') RETURNING id", uuid.New(), ownerID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %v", err)
	}

	req := models.AddMemberRequest{
		Email: &memberEmail,
	}

	rr := makeRequest(router, "POST", "/api/my/groups/"+groupID+"/members", req, ownerToken)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp models.GroupMember
	parseResponse(t, rr, &resp)

	if resp.Member.ID != memberID {
		t.Errorf("expected member ID %s, got %s", memberID, resp.Member.ID)
	}
	if resp.GroupID != groupID {
		t.Errorf("expected group ID %s, got %s", groupID, resp.GroupID)
	}
}

func TestGroupsAddMember_ByUserID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create owner user
	ownerEmail := "groupowner6@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Group Owner")
	ownerToken := getAuthToken(ownerID, ownerEmail)

	// Create user to add as member
	memberEmail := "membertoadd2@example.com"
	memberID := createTestUser(t, db, memberEmail, "password123", "Member to Add")

	// Insert a group
	var groupID string
	err := db.QueryRowContext(ctx, "INSERT INTO visibility_groups (id, owner_id, name, visibility_level) VALUES (?, ?, 'Group', 'work') RETURNING id", uuid.New(), ownerID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %v", err)
	}

	req := models.AddMemberRequest{
		UserID: &memberID,
	}

	rr := makeRequest(router, "POST", "/api/my/groups/"+groupID+"/members", req, ownerToken)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp models.GroupMember
	parseResponse(t, rr, &resp)

	if resp.Member.ID != memberID {
		t.Errorf("expected member ID %s, got %s", memberID, resp.Member.ID)
	}
}

func TestGroupsAddMember_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create owner user
	ownerEmail := "groupowner7@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Group Owner")
	ownerToken := getAuthToken(ownerID, ownerEmail)

	// Insert a group
	var groupID string
	err := db.QueryRowContext(ctx, "INSERT INTO visibility_groups (id, owner_id, name, visibility_level) VALUES (?, ?, 'Group', 'work') RETURNING id", uuid.New(), ownerID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %v", err)
	}

	nonExistentEmail := "nonexistent@example.com"
	req := models.AddMemberRequest{
		Email: &nonExistentEmail,
	}

	rr := makeRequest(router, "POST", "/api/my/groups/"+groupID+"/members", req, ownerToken)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}

	errResp := parseErrorResponse(t, rr)
	if errResp.Error != "user not found" {
		t.Errorf("unexpected error message: %s", errResp.Error)
	}
}

func TestGroupsAddMember_MissingEmailAndUserID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create owner user
	ownerEmail := "groupowner8@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Group Owner")
	ownerToken := getAuthToken(ownerID, ownerEmail)

	// Insert a group
	var groupID string
	err := db.QueryRowContext(ctx, "INSERT INTO visibility_groups (id, owner_id, name, visibility_level) VALUES (?, ?, 'Group', 'work') RETURNING id", uuid.New(), ownerID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %v", err)
	}

	req := models.AddMemberRequest{}

	rr := makeRequest(router, "POST", "/api/my/groups/"+groupID+"/members", req, ownerToken)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}

	errResp := parseErrorResponse(t, rr)
	if errResp.Error != "either email or userId is required" {
		t.Errorf("unexpected error message: %s", errResp.Error)
	}
}

func TestGroupsAddMember_NotOwner(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create two users
	ownerEmail := "groupowner9@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Group Owner")

	otherEmail := "groupother4@example.com"
	otherID := createTestUser(t, db, otherEmail, "password123", "Other User")
	otherToken := getAuthToken(otherID, otherEmail)

	// Create user to add
	memberEmail := "member@example.com"
	_ = createTestUser(t, db, memberEmail, "password123", "Member")

	// Insert a group for owner
	var groupID string
	err := db.QueryRowContext(ctx, "INSERT INTO visibility_groups (id, owner_id, name, visibility_level) VALUES (?, ?, 'Private Group', 'family') RETURNING id", uuid.New(), ownerID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %v", err)
	}

	req := models.AddMemberRequest{
		Email: &memberEmail,
	}

	rr := makeRequest(router, "POST", "/api/my/groups/"+groupID+"/members", req, otherToken)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rr.Code)
	}

	errResp := parseErrorResponse(t, rr)
	if errResp.Error != "you don't have access to this group" {
		t.Errorf("unexpected error message: %s", errResp.Error)
	}
}

func TestGroupsAddMember_DuplicateMember(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create owner user
	ownerEmail := "groupowner10@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Group Owner")
	ownerToken := getAuthToken(ownerID, ownerEmail)

	// Create member user
	memberEmail := "duplicatemember@example.com"
	memberID := createTestUser(t, db, memberEmail, "password123", "Duplicate Member")

	// Insert a group
	var groupID string
	err := db.QueryRowContext(ctx, "INSERT INTO visibility_groups (id, owner_id, name, visibility_level) VALUES (?, ?, 'Group', 'family') RETURNING id", uuid.New(), ownerID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %v", err)
	}

	// Add member first time
	_, err = db.ExecContext(ctx, "INSERT INTO group_members (id, group_id, member_id, added_by) VALUES (?, ?, ?, ?)", uuid.New(), groupID, memberID, ownerID)
	if err != nil {
		t.Fatalf("failed to add member: %v", err)
	}

	// Try to add same member again
	req := models.AddMemberRequest{
		Email: &memberEmail,
	}

	rr := makeRequest(router, "POST", "/api/my/groups/"+groupID+"/members", req, ownerToken)

	if rr.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rr.Code)
	}

	errResp := parseErrorResponse(t, rr)
	if errResp.Error != "user is already a member of this group" {
		t.Errorf("unexpected error message: %s", errResp.Error)
	}
}

func TestGroupsRemoveMember_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create owner user
	ownerEmail := "groupowner11@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Group Owner")
	ownerToken := getAuthToken(ownerID, ownerEmail)

	// Create member user
	memberEmail := "membertoremove@example.com"
	memberID := createTestUser(t, db, memberEmail, "password123", "Member to Remove")

	// Insert a group
	var groupID string
	err := db.QueryRowContext(ctx, "INSERT INTO visibility_groups (id, owner_id, name, visibility_level) VALUES (?, ?, 'Group', 'friends') RETURNING id", uuid.New(), ownerID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %v", err)
	}

	// Add member and get the membership ID
	var memberRecordID string
	err = db.QueryRowContext(ctx, "INSERT INTO group_members (id, group_id, member_id, added_by) VALUES (?, ?, ?, ?) RETURNING id", uuid.New(), groupID, memberID, ownerID).Scan(&memberRecordID)
	if err != nil {
		t.Fatalf("failed to add member: %v", err)
	}

	rr := makeRequest(router, "DELETE", "/api/my/groups/"+groupID+"/members/"+memberRecordID, nil, ownerToken)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestGroupsRemoveMember_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create owner user
	ownerEmail := "groupowner12@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Group Owner")
	ownerToken := getAuthToken(ownerID, ownerEmail)

	// Insert a group
	var groupID string
	err := db.QueryRowContext(ctx, "INSERT INTO visibility_groups (id, owner_id, name, visibility_level) VALUES (?, ?, 'Group', 'friends') RETURNING id", uuid.New(), ownerID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %v", err)
	}

	rr := makeRequest(router, "DELETE", "/api/my/groups/"+groupID+"/members/nonexistent-id", nil, ownerToken)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}

	errResp := parseErrorResponse(t, rr)
	if errResp.Error != "member not found" {
		t.Errorf("unexpected error message: %s", errResp.Error)
	}
}

func TestGroupsRemoveMember_NotOwner(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupTestData(t, db)

	router := NewRouter(db)
	ctx := context.Background()

	// Create two users
	ownerEmail := "groupowner13@example.com"
	ownerID := createTestUser(t, db, ownerEmail, "password123", "Group Owner")

	otherEmail := "groupother5@example.com"
	otherID := createTestUser(t, db, otherEmail, "password123", "Other User")
	otherToken := getAuthToken(otherID, otherEmail)

	// Create member user
	memberEmail := "member5@example.com"
	memberID := createTestUser(t, db, memberEmail, "password123", "Member")

	// Insert a group for owner
	var groupID string
	err := db.QueryRowContext(ctx, "INSERT INTO visibility_groups (id, owner_id, name, visibility_level) VALUES (?, ?, 'Private Group', 'family') RETURNING id", uuid.New(), ownerID).Scan(&groupID)
	if err != nil {
		t.Fatalf("failed to insert group: %v", err)
	}

	// Add member
	var memberRecordID string
	err = db.QueryRowContext(ctx, "INSERT INTO group_members (id, group_id, member_id, added_by) VALUES (?, ?, ?, ?) RETURNING id", uuid.New(), groupID, memberID, ownerID).Scan(&memberRecordID)
	if err != nil {
		t.Fatalf("failed to add member: %v", err)
	}

	rr := makeRequest(router, "DELETE", "/api/my/groups/"+groupID+"/members/"+memberRecordID, nil, otherToken)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rr.Code)
	}

	errResp := parseErrorResponse(t, rr)
	if errResp.Error != "you don't have access to this group" {
		t.Errorf("unexpected error message: %s", errResp.Error)
	}
}
