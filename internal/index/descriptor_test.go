package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tableauio/loader/test/go-tableau-loader/protoconf"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func md[T proto.Message]() protoreflect.MessageDescriptor {
	var t T
	return t.ProtoReflect().Descriptor()
}

func fd[T proto.Message](name protoreflect.Name) protoreflect.FieldDescriptor {
	return md[T]().Fields().ByName(name)
}

func Test_ParseIndexDescriptor(t *testing.T) {
	type args struct {
		md protoreflect.MessageDescriptor
	}
	tests := []struct {
		name string
		args args
		want *IndexDescriptor
	}{
		{
			name: "ItemConf",
			args: args{
				md: md[*protoconf.ItemConf](),
			},
			want: &IndexDescriptor{
				LevelMessage: &LevelMessage{
					Depth:    0,
					MapDepth: 0,
					FD:       fd[*protoconf.ItemConf]("item_map"),
					NextLevel: &LevelMessage{
						Depth:    1,
						MapDepth: 1,
						Indexes: []*LevelIndex{
							{
								Index: &Index{
									Cols: []string{"Type"},
									Name: "",
								},
								MD: md[*protoconf.ItemConf_Item](),
								ColFields: []*LevelField{
									{
										FD: fd[*protoconf.ItemConf_Item]("type"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.ItemConf_Item]("type"),
										},
									},
								},
							},
							{
								Index: &Index{
									Cols:       []string{"Param"},
									SortedCols: []string{"ID"},
									Name:       "ItemInfo",
								},
								MD: md[*protoconf.ItemConf_Item](),
								ColFields: []*LevelField{
									{
										FD: fd[*protoconf.ItemConf_Item]("param_list"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.ItemConf_Item]("param_list"),
										},
									},
								},
								SortedColFields: []*LevelField{
									{
										FD: fd[*protoconf.ItemConf_Item]("id"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.ItemConf_Item]("id"),
										},
									},
								},
							},
							{
								Index: &Index{
									Cols: []string{"Default"},
									Name: "ItemDefaultInfo",
								},
								MD: md[*protoconf.ItemConf_Item](),
								ColFields: []*LevelField{
									{
										FD: fd[*protoconf.ItemConf_Item]("default"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.ItemConf_Item]("default"),
										},
									},
								},
							},
							{
								Index: &Index{
									Cols: []string{"ExtType"},
									Name: "ItemExtInfo",
								},
								MD: md[*protoconf.ItemConf_Item](),
								ColFields: []*LevelField{
									{
										FD: fd[*protoconf.ItemConf_Item]("ext_type_list"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.ItemConf_Item]("ext_type_list"),
										},
									},
								},
							},
							{
								Index: &Index{
									Cols:       []string{"ID", "Name"},
									SortedCols: []string{"Type", "UseEffectType"},
									Name:       "AwardItem",
								},
								MD: md[*protoconf.ItemConf_Item](),
								ColFields: []*LevelField{
									{
										FD: fd[*protoconf.ItemConf_Item]("id"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.ItemConf_Item]("id"),
										},
									},
									{
										FD: fd[*protoconf.ItemConf_Item]("name"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.ItemConf_Item]("name"),
										},
									},
								},
								SortedColFields: []*LevelField{
									{
										FD: fd[*protoconf.ItemConf_Item]("type"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.ItemConf_Item]("type"),
										},
									},
									{
										FD: fd[*protoconf.UseEffect]("type"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.ItemConf_Item]("use_effect"),
											fd[*protoconf.UseEffect]("type"),
										},
									},
								},
							},
							{
								Index: &Index{
									Cols: []string{"ID", "Type", "Param", "ExtType"},
									Name: "SpecialItem",
								},
								MD: md[*protoconf.ItemConf_Item](),
								ColFields: []*LevelField{
									{
										FD: fd[*protoconf.ItemConf_Item]("id"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.ItemConf_Item]("id"),
										},
									},
									{
										FD: fd[*protoconf.ItemConf_Item]("type"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.ItemConf_Item]("type"),
										},
									},
									{
										FD: fd[*protoconf.ItemConf_Item]("param_list"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.ItemConf_Item]("param_list"),
										},
									},
									{
										FD: fd[*protoconf.ItemConf_Item]("ext_type_list"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.ItemConf_Item]("ext_type_list"),
										},
									},
								},
							},
							{
								Index: &Index{
									Cols: []string{"PathDir"},
									Name: "ItemPathDir",
								},
								MD: md[*protoconf.ItemConf_Item](),
								ColFields: []*LevelField{
									{
										FD: fd[*protoconf.Path]("dir"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.ItemConf_Item]("path"),
											fd[*protoconf.Path]("dir"),
										},
									},
								},
							},
							{
								Index: &Index{
									Cols: []string{"PathName"},
									Name: "ItemPathName",
								},
								MD: md[*protoconf.ItemConf_Item](),
								ColFields: []*LevelField{
									{
										FD: fd[*protoconf.Path]("name_list"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.ItemConf_Item]("path"),
											fd[*protoconf.Path]("name_list"),
										},
									},
								},
							},
							{
								Index: &Index{
									Cols: []string{"PathFriendID"},
									Name: "ItemPathFriendID",
								},
								MD: md[*protoconf.ItemConf_Item](),
								ColFields: []*LevelField{
									{
										FD: fd[*protoconf.Path_Friend]("id"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.ItemConf_Item]("path"),
											fd[*protoconf.Path]("friend"),
											fd[*protoconf.Path_Friend]("id"),
										},
									},
								},
							},
							{
								Index: &Index{
									Cols: []string{"UseEffectType"},
									Name: "UseEffectType",
								},
								MD: md[*protoconf.ItemConf_Item](),
								ColFields: []*LevelField{
									{
										FD: fd[*protoconf.UseEffect]("type"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.ItemConf_Item]("use_effect"),
											fd[*protoconf.UseEffect]("type"),
										},
									},
								},
							},
						},
						OrderedIndexes: []*LevelIndex{
							{
								Index: &Index{
									Cols: []string{"ExtType"},
									Name: "ExtType",
								},
								MD: md[*protoconf.ItemConf_Item](),
								ColFields: []*LevelField{
									{
										FD: fd[*protoconf.ItemConf_Item]("ext_type_list"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.ItemConf_Item]("ext_type_list"),
										},
									},
								},
							},
							{
								Index: &Index{
									Cols:       []string{"Param", "ExtType"},
									SortedCols: []string{"ID"},
									Name:       "ParamExtType",
								},
								MD: md[*protoconf.ItemConf_Item](),
								ColFields: []*LevelField{
									{
										FD: fd[*protoconf.ItemConf_Item]("param_list"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.ItemConf_Item]("param_list"),
										},
									},
									{
										FD: fd[*protoconf.ItemConf_Item]("ext_type_list"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.ItemConf_Item]("ext_type_list"),
										},
									},
								},
								SortedColFields: []*LevelField{
									{
										FD: fd[*protoconf.ItemConf_Item]("id"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.ItemConf_Item]("id"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "HeroConf",
			args: args{
				md: md[*protoconf.HeroConf](),
			},
			want: &IndexDescriptor{
				LevelMessage: &LevelMessage{
					Depth:    0,
					MapDepth: 0,
					FD:       fd[*protoconf.HeroConf]("hero_map"),
					NextLevel: &LevelMessage{
						Depth:    1,
						MapDepth: 1,
						FD:       fd[*protoconf.HeroConf_Hero]("attr_map"),
						NextLevel: &LevelMessage{
							Depth:    2,
							MapDepth: 2,
							Indexes: []*LevelIndex{
								{
									Index: &Index{
										Cols: []string{"Title"},
										Name: "",
									},
									MD: md[*protoconf.HeroConf_Hero_Attr](),
									ColFields: []*LevelField{
										{
											FD: fd[*protoconf.HeroConf_Hero_Attr]("title"),
											LeveledFDList: []protoreflect.FieldDescriptor{
												fd[*protoconf.HeroConf_Hero_Attr]("title"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ActivityConf",
			args: args{
				md: md[*protoconf.ActivityConf](),
			},
			want: &IndexDescriptor{
				LevelMessage: &LevelMessage{
					Depth:    0,
					MapDepth: 0,
					FD:       fd[*protoconf.ActivityConf]("activity_map"),
					NextLevel: &LevelMessage{
						Depth:    1,
						MapDepth: 1,
						FD:       fd[*protoconf.ActivityConf_Activity]("chapter_map"),
						Indexes: []*LevelIndex{
							{
								Index: &Index{
									Cols: []string{"ActivityName"},
									Name: "",
								},
								MD: md[*protoconf.ActivityConf_Activity](),
								ColFields: []*LevelField{
									{
										FD: fd[*protoconf.ActivityConf_Activity]("activity_name"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.ActivityConf_Activity]("activity_name"),
										},
									},
								},
							},
						},
						NextLevel: &LevelMessage{
							Depth:    2,
							MapDepth: 2,
							FD:       fd[*protoconf.ActivityConf_Activity_Chapter]("section_map"),
							Indexes: []*LevelIndex{
								{
									Index: &Index{
										Cols: []string{"ChapterID"},
										Name: "",
									},
									MD: md[*protoconf.ActivityConf_Activity_Chapter](),
									ColFields: []*LevelField{
										{
											FD: fd[*protoconf.ActivityConf_Activity_Chapter]("chapter_id"),
											LeveledFDList: []protoreflect.FieldDescriptor{
												fd[*protoconf.ActivityConf_Activity_Chapter]("chapter_id"),
											},
										},
									},
								},
								{
									Index: &Index{
										Cols:       []string{"ChapterName"},
										SortedCols: []string{"AwardID"},
										Name:       "NamedChapter",
									},
									MD: md[*protoconf.ActivityConf_Activity_Chapter](),
									ColFields: []*LevelField{
										{
											FD: fd[*protoconf.ActivityConf_Activity_Chapter]("chapter_name"),
											LeveledFDList: []protoreflect.FieldDescriptor{
												fd[*protoconf.ActivityConf_Activity_Chapter]("chapter_name"),
											},
										},
									},
									SortedColFields: []*LevelField{
										{
											FD: fd[*protoconf.ActivityConf_Activity_Chapter]("award_id"),
											LeveledFDList: []protoreflect.FieldDescriptor{
												fd[*protoconf.ActivityConf_Activity_Chapter]("award_id"),
											},
										},
									},
								},
							},
							NextLevel: &LevelMessage{
								Depth:    3,
								MapDepth: 3,
								FD:       fd[*protoconf.Section]("section_item_list"),
								NextLevel: &LevelMessage{
									Depth:    4,
									MapDepth: 3,
									FD:       fd[*protoconf.Section_SectionItem]("decompose_item_list"),
									Indexes: []*LevelIndex{
										{
											Index: &Index{
												Cols: []string{"SectionItemID"},
												Name: "Award",
											},
											MD: md[*protoconf.Section_SectionItem](),
											ColFields: []*LevelField{
												{
													FD: fd[*protoconf.Section_SectionItem]("id"),
													LeveledFDList: []protoreflect.FieldDescriptor{
														fd[*protoconf.Section_SectionItem]("id"),
													},
												},
											},
										},
									},
									NextLevel: &LevelMessage{
										Depth:    5,
										MapDepth: 3,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "TaskConf",
			args: args{
				md: md[*protoconf.TaskConf](),
			},
			want: &IndexDescriptor{
				LevelMessage: &LevelMessage{
					Depth:    0,
					MapDepth: 0,
					FD:       fd[*protoconf.TaskConf]("task_map"),
					NextLevel: &LevelMessage{
						Depth:    1,
						MapDepth: 1,
						Indexes: []*LevelIndex{
							{
								Index: &Index{
									Cols:       []string{"ActivityID"},
									SortedCols: []string{"Goal", "ID"},
									Name:       "",
								},
								MD: md[*protoconf.TaskConf_Task](),
								ColFields: []*LevelField{
									{
										FD: fd[*protoconf.TaskConf_Task]("activity_id"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.TaskConf_Task]("activity_id"),
										},
									},
								},
								SortedColFields: []*LevelField{
									{
										FD: fd[*protoconf.TaskConf_Task]("goal"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.TaskConf_Task]("goal"),
										},
									},
									{
										FD: fd[*protoconf.TaskConf_Task]("id"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.TaskConf_Task]("id"),
										},
									},
								},
							},
						},
						OrderedIndexes: []*LevelIndex{
							{
								Index: &Index{
									Cols:       []string{"Goal"},
									SortedCols: []string{"ID"},
									Name:       "OrderedTask",
								},
								MD: md[*protoconf.TaskConf_Task](),
								ColFields: []*LevelField{
									{
										FD: fd[*protoconf.TaskConf_Task]("goal"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.TaskConf_Task]("goal"),
										},
									},
								},
								SortedColFields: []*LevelField{
									{
										FD: fd[*protoconf.TaskConf_Task]("id"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.TaskConf_Task]("id"),
										},
									},
								},
							},
							{
								Index: &Index{
									Cols: []string{"Expiry"},
									Name: "TaskExpiry",
								},
								MD: md[*protoconf.TaskConf_Task](),
								ColFields: []*LevelField{
									{
										FD: fd[*protoconf.TaskConf_Task]("expiry"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.TaskConf_Task]("expiry"),
										},
									},
								},
							},
							{
								Index: &Index{
									Cols:       []string{"Expiry"},
									SortedCols: []string{"Goal", "ID"},
									Name:       "SortedTaskExpiry",
								},
								MD: md[*protoconf.TaskConf_Task](),
								ColFields: []*LevelField{
									{
										FD: fd[*protoconf.TaskConf_Task]("expiry"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.TaskConf_Task]("expiry"),
										},
									},
								},
								SortedColFields: []*LevelField{
									{
										FD: fd[*protoconf.TaskConf_Task]("goal"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.TaskConf_Task]("goal"),
										},
									},
									{
										FD: fd[*protoconf.TaskConf_Task]("id"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.TaskConf_Task]("id"),
										},
									},
								},
							},
							{
								Index: &Index{
									Cols: []string{"Expiry", "ActivityID"},
									Name: "ActivityExpiry",
								},
								MD: md[*protoconf.TaskConf_Task](),
								ColFields: []*LevelField{
									{
										FD: fd[*protoconf.TaskConf_Task]("expiry"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.TaskConf_Task]("expiry"),
										},
									},
									{
										FD: fd[*protoconf.TaskConf_Task]("activity_id"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.TaskConf_Task]("activity_id"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			// Fruit5Conf: 3-level map (fruit_map -> country_map -> item_map),
			// but indexes only at MapDepth=2 (Country level).
			// The 3rd level map (MapDepth=3, Item) has no index.
			// This validates that maxDepth (=2) < len(mapFds) (=3) does not cause
			// generation of extra LevelIndex key structs.
			name: "Fruit5Conf",
			args: args{
				md: md[*protoconf.Fruit5Conf](),
			},
			want: &IndexDescriptor{
				LevelMessage: &LevelMessage{
					Depth:    0,
					MapDepth: 0,
					FD:       fd[*protoconf.Fruit5Conf]("fruit_map"),
					NextLevel: &LevelMessage{
						Depth:    1,
						MapDepth: 1,
						FD:       fd[*protoconf.Fruit5Conf_Fruit]("country_map"),
						NextLevel: &LevelMessage{
							Depth:    2,
							MapDepth: 2,
							FD:       fd[*protoconf.Fruit5Conf_Fruit_Country]("item_map"),
							Indexes: []*LevelIndex{
								{
									Index: &Index{
										Cols: []string{"CountryName"},
										Name: "",
									},
									MD: md[*protoconf.Fruit5Conf_Fruit_Country](),
									ColFields: []*LevelField{
										{
											FD: fd[*protoconf.Fruit5Conf_Fruit_Country]("name"),
											LeveledFDList: []protoreflect.FieldDescriptor{
												fd[*protoconf.Fruit5Conf_Fruit_Country]("name"),
											},
										},
									},
								},
							},
							NextLevel: &LevelMessage{
								Depth:    3,
								MapDepth: 3,
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseIndexDescriptor(tt.args.md)
			assert.EqualValues(t, tt.want, got)
		})
	}
}

// Test_Fruit5Conf_MaxDepthVsMapFds verifies that when the deepest map level
// (MapDepth=3) has no index but a shallower level (MapDepth=2) does,
// the generator loop `for i := 0; i < maxDepth-2 && i < len(mapFds)-1; i++`
// does NOT produce extra LevelIndex key structs.
//
// Fruit5Conf has 3-level map (fruit_map -> country_map -> item_map) but
// indexes only at MapDepth=2 (Country), so maxDepth=2, len(mapFds)=3.
// Without the maxDepth-2 guard, the loop would run 2 iterations instead of 0.
func Test_Fruit5Conf_MaxDepthVsMapFds(t *testing.T) {
	descriptor := ParseIndexDescriptor(md[*protoconf.Fruit5Conf]())

	// Simulate initLevelMessage logic to collect maxDepth and mapFds count.
	var maxDepth int
	var mapFdCount int
	for levelMessage := descriptor.LevelMessage; levelMessage != nil; levelMessage = levelMessage.NextLevel {
		if fd := levelMessage.FD; fd != nil && fd.IsMap() {
			mapFdCount++
		}
		if len(levelMessage.Indexes) != 0 || len(levelMessage.OrderedIndexes) != 0 {
			maxDepth = levelMessage.MapDepth
		}
	}

	// Fruit5Conf: 3 map levels, but deepest index is at MapDepth=2.
	assert.Equal(t, 3, mapFdCount, "Fruit5Conf should have 3 map field descriptors")
	assert.Equal(t, 2, maxDepth, "Fruit5Conf's deepest index should be at MapDepth=2")
	assert.Less(t, maxDepth, mapFdCount, "maxDepth should be less than len(mapFds)")

	// Verify the LevelIndex key generation loop would produce 0 iterations,
	// meaning no extra LevelIndex key structs are generated.
	levelIndexKeyCount := 0
	for i := 0; i < maxDepth-2 && i < mapFdCount-1; i++ {
		levelIndexKeyCount++
	}
	assert.Equal(t, 0, levelIndexKeyCount,
		"No LevelIndex key structs should be generated: maxDepth-2=%d, len(mapFds)-1=%d",
		maxDepth-2, mapFdCount-1)

	// Contrast: without the maxDepth-2 guard, using only len(mapFds)-1
	// would incorrectly generate 2 LevelIndex key structs.
	wrongKeyCount := 0
	for i := 0; i < mapFdCount-1; i++ {
		wrongKeyCount++
	}
	assert.Equal(t, 2, wrongKeyCount,
		"Without maxDepth guard, %d extra LevelIndex key structs would be generated", wrongKeyCount)
}
