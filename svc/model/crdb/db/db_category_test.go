package db

import (
	"context"
	"fmt"
	"sort"
	"testing"
)

func TestCategoriesCRUD(t *testing.T) {
	ctx := context.Background()
	db, stop := dut(ctx, t)
	defer stop()

	// Test fake category
	_, err := db.Category().Get(ctx, "foo")
	if want, got := CategoryErrorNotFound, err; want != got {
		t.Fatalf("Category.Get(invalid uuid): wanted %q, got %q", want, got)
	}

	// Test root category
	_, err = db.Category().Get(ctx, RootCategory)
	if err != nil {
		t.Fatalf("Category.Get(root): %v", err)
	}

	// Check nonexistant category.
	_, err = db.Category().Get(ctx, "ddacd7d8-6d4e-4013-be58-cac97fe12cc6")
	if want, got := CategoryErrorNotFound, err; want != got {
		t.Fatalf("Category.Get(nonexistent): wanted %q, got %q", want, got)
	}

	// Test creation happy path
	cat, err := db.Category().New(ctx, &Category{
		Name:       "test",
		ParentUUID: RootCategory,
	})
	if err != nil {
		t.Fatalf("Could not create category: %v", err)
	}
	if cat.UUID == "" {
		t.Fatalf("New category does not have UUID")
	}
	testCat := cat // used later in tests

	// Test an unparseable parent
	_, err = db.Category().New(ctx, &Category{
		Name:       "test",
		ParentUUID: "fake", // unparseable
	})
	if got, want := err, CategoryErrorParentNotFound; want != got {
		t.Fatalf("Category.New(invalid parent uuid): wanted %q, got %q", want, got)
	}

	// Test a nonexistent parent
	_, err = db.Category().New(ctx, &Category{
		Name:       "test",
		ParentUUID: "ddacd7d8-6d4e-4013-be58-cac97fe12cc6", // does not exist
	})
	if got, want := err, CategoryErrorParentNotFound; want != got {
		t.Fatalf("Category.New(nonexistent parent): wanted %q, got %q", want, got)
	}

	// Test a duplicate name
	_, err = db.Category().New(ctx, &Category{
		Name:       "test",
		ParentUUID: RootCategory,
	})
	if got, want := err, CategoryErrorDuplicateName; want != got {
		t.Fatalf("Category.New(nonexistent parent): wanted %q, got %q", want, got)
	}

	// Add a bunch of categories
	categories := make([]string, 128)
	for i, _ := range categories {
		name := fmt.Sprintf("category %d", i)
		description := fmt.Sprintf("yes this is the category number %d very special", i)

		cat, err = db.Category().New(ctx, &Category{
			Name:        name,
			Description: description,
			ParentUUID:  RootCategory,
		})
		if err != nil {
			t.Fatalf("Could not create category %q: %v", name, err)
		}
		categories[i] = cat.UUID
	}

	// Check retrieval of categories
	for i, uuid := range categories {
		wantName := fmt.Sprintf("category %d", i)
		wantDescription := fmt.Sprintf("yes this is the category number %d very special", i)

		cat, err = db.Category().Get(ctx, uuid)
		if err != nil {
			t.Fatalf("Could not retrieve category %q: %v", wantName, err)
		}
		if want, got := wantName, cat.Name; want != got {
			t.Fatalf("Category name %q got saved as %q", want, got)
		}
		if want, got := wantDescription, cat.Description; want != got {
			t.Fatalf("Category description %q got saved as %q", want, got)
		}
		if want, got := RootCategory, cat.ParentUUID; want != got {
			t.Fatalf("Category parent UUID %q got saved as %q", want, got)
		}
	}

	// Check updating a category
	cat, err = db.Category().New(ctx, &Category{
		Name:        "foo",
		Description: "bar",
		ParentUUID:  testCat.UUID,
	})
	if err != nil {
		t.Fatalf("Could not create category: %v", err)
	}
	testCat2 := cat // used later in tests

	cat.Name += "!"
	cat.Description += "!"
	err = db.Category().Update(ctx, cat)
	if err != nil {
		t.Fatalf("Could not update category: %v", err)
	}

	cat, err = db.Category().Get(ctx, cat.UUID)
	if err != nil {
		t.Fatalf("Could not retrive category: %v", err)
	}

	if want, got := "foo!", cat.Name; want != got {
		t.Errorf("Updated category name: wanted %q, got %q", want, got)
	}
	if want, got := "bar!", cat.Description; want != got {
		t.Errorf("Updated category description: wanted %q, got %q", want, got)
	}

	// Check updating category with invalid args
	cat.Name = ""
	err = db.Category().Update(ctx, cat)
	if err == nil {
		t.Fatalf("Category with empty name didn't get rejected")
	}

	// Check updating category to nonexistent parent
	cat.Name = "foo!"
	cat.ParentUUID = "f1e89031-3307-4c47-b791-76136e13fc58" // doesn't exist
	err = db.Category().Update(ctx, cat)
	if want, got := CategoryErrorParentNotFound, err; want != got {
		t.Errorf("Category.Update(nonexistent parent): wanted %q got %q", want, got)
	}

	// Check updating category to invalid parent
	cat.ParentUUID = "foo"
	err = db.Category().Update(ctx, cat)
	if want, got := CategoryErrorParentNotFound, err; want != got {
		t.Errorf("Category.Update(invalid parent): wanted %q got %q", want, got)
	}

	// Check updating category to duplicate name.
	cat.ParentUUID = RootCategory
	cat.Name = "test"
	err = db.Category().Update(ctx, cat)
	if want, got := CategoryErrorDuplicateName, err; want != got {
		t.Errorf("Category.Update(duplicate name): wanted %q got %q", want, got)
	}

	// Move a category to another parent.
	cat, err = db.Category().New(ctx, &Category{
		Name:        "foo",
		Description: "bar",
		ParentUUID:  testCat2.UUID,
	})
	if err != nil {
		t.Fatalf("Could not create category: %v", err)
	}
	cat.ParentUUID = testCat.UUID
	err = db.Category().Update(ctx, cat)
	if err != nil {
		t.Fatalf("Could not update category: %v", err)
	}

	cat, err = db.Category().Get(ctx, cat.UUID)
	if err != nil {
		t.Fatalf("Could not retrive category: %v", err)
	}

	if want, got := testCat.UUID, cat.ParentUUID; want != got {
		t.Fatalf("Updated category parent: wanted %q, got %q", want, got)
	}

	// Check removing root category
	err = db.Category().Delete(ctx, RootCategory)
	if want, got := CategoryErrorCannotDeleteRoot, err; want != got {
		t.Fatalf("Category.Delete(root): wanted %q, got %q", want, got)
	}

	// Check removing a non-leaf category
	err = db.Category().Delete(ctx, testCat.UUID)
	if want, got := CategoryErrorNotEmpty, err; want != got {
		t.Fatalf("Category.Delete(non-leaf): wanted %q, got %q", want, got)
	}

	// Check removing a leaf category
	err = db.Category().Delete(ctx, cat.UUID)
	if err != nil {
		t.Fatalf("Category.Delete(leaf): %v", err)
	}
	_, err = db.Category().Get(ctx, cat.UUID)
	if want, got := CategoryErrorNotFound, err; want != got {
		t.Fatalf("Category.Get(removed):wanted %q, got %q", want, got)
	}
}

func TestCategoriesTree(t *testing.T) {
	ctx := context.Background()
	db, stop := dut(ctx, t)
	defer stop()

	mkNode := func(name, parent string) string {
		cat, err := db.Category().New(ctx, &Category{
			ParentUUID: parent,
			Name:       name,
		})
		if err != nil {
			t.Fatalf("Could not create category %q: %v", name, err)
		}
		return cat.UUID
	}

	// root -.-> A -.-> 00
	//       |      |-> ..
	//       |      '-> 49
	//       |
	//       |-> B -.-> 100 ---> 200
	//       |      |-> 101 ---> 201
	//       |      '-> 102 -.-> 201 --> 300
	//       |               '-> 202
	//       '-> C

	// Make tree
	a := mkNode("A", RootCategory)
	for i := 0; i <= 49; i += 1 {
		mkNode(fmt.Sprintf("%02d", i), a)
	}

	b := mkNode("B", RootCategory)
	n100 := mkNode("100", b)
	mkNode("200", n100)
	n101 := mkNode("101", b)
	mkNode("201", n101)
	n102 := mkNode("102", b)
	n201 := mkNode("201", n102)
	mkNode("300", n201)
	mkNode("202", n102)

	_ = mkNode("C", RootCategory)

	ensureChildren := func(node *CategoryNode, children []string) {
		if want, got := len(children), len(node.Children); want != got {
			t.Fatalf("node %q has %d children, want %d", node.Name, got, want)
		}

		sort.Slice(node.Children, func(i, j int) bool { return node.Children[i].Name < node.Children[j].Name })

		for i, want := range children {
			if got := node.Children[i].Name; want != got {
				t.Fatalf("node %q has child %d %q, want %q", node.Name, i, got, want)
			}
		}
	}

	// Get entire tree and ensure the gang's all there.
	func() {
		tree, err := db.Category().GetTree(ctx, RootCategory, 4)
		if err != nil {
			t.Fatalf("GetTree(root): %v", err)
		}
		ensureChildren(tree, []string{"A", "B", "C"})
		gotA, gotB, gotC := tree.Children[0], tree.Children[1], tree.Children[2]

		// 0 .. 49 under A
		want := make([]string, 50)
		for i, _ := range want {
			want[i] = fmt.Sprintf("%02d", i)
		}
		ensureChildren(gotA, want)

		// 100 .. 102 under B
		ensureChildren(gotB, []string{"100", "101", "102"})
		got100, got101, got102 := gotB.Children[0], gotB.Children[1], gotB.Children[2]

		// 200 under 100
		ensureChildren(got100, []string{"200"})
		// 201 under 101
		ensureChildren(got101, []string{"201"})

		// 201 (repeat) and 202 under 102
		ensureChildren(got102, []string{"201", "202"})
		got201 := got102.Children[0]

		// 300 under 201
		ensureChildren(got201, []string{"300"})

		// nothing under C
		ensureChildren(gotC, []string{})
	}()

	// Get tree under B and ensure the gang's all here.
	func() {
		gotB, err := db.Category().GetTree(ctx, b, 3)
		if err != nil {
			t.Fatalf("GetTree(B, 3): %v", err)
		}
		// 100 .. 102 under B
		ensureChildren(gotB, []string{"100", "101", "102"})
		got100, got101, got102 := gotB.Children[0], gotB.Children[1], gotB.Children[2]

		// 200 under 100
		ensureChildren(got100, []string{"200"})
		// 201 under 101
		ensureChildren(got101, []string{"201"})

		// 201 (repeat) and 202 under 102
		ensureChildren(got102, []string{"201", "202"})
		got201 := got102.Children[0]

		// 300 under 201
		ensureChildren(got201, []string{"300"})
	}()

	// Get only direct children under B, make sure only they are returned.
	func() {
		gotB, err := db.Category().GetTree(ctx, b, 1)
		if err != nil {
			t.Fatalf("GetTree(B, 1): %v", err)
		}
		// 100 .. 102 under B
		ensureChildren(gotB, []string{"100", "101", "102"})
		got100, got101, got102 := gotB.Children[0], gotB.Children[1], gotB.Children[2]

		// nothing under 100
		ensureChildren(got100, []string{})
		// nothing under 101
		ensureChildren(got101, []string{})
		// nothing under 102
		ensureChildren(got102, []string{})
	}()

	// Get only B, ensure _no_ children are returned.
	func() {
		gotB, err := db.Category().GetTree(ctx, b, 0)
		if err != nil {
			t.Fatalf("GetTree(B, 0): %v", err)
		}
		ensureChildren(gotB, []string{})
	}()
}
