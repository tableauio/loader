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
					FD:       fd[*protoconf.ItemConf]("item_map"),
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
		{
			name: "HeroConf",
			args: args{
				md: md[*protoconf.HeroConf](),
			},
			want: &IndexDescriptor{
				LevelMessage: &LevelMessage{
					FD:       fd[*protoconf.HeroConf]("hero_map"),
					Depth:    1,
					MapDepth: 1,
					NextLevel: &LevelMessage{
						FD:       fd[*protoconf.HeroConf_Hero]("attr_map"),
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
		{
			name: "ActivityConf",
			args: args{
				md: md[*protoconf.ActivityConf](),
			},
			want: &IndexDescriptor{
				LevelMessage: &LevelMessage{
					FD:       fd[*protoconf.ActivityConf]("activity_map"),
					Depth:    1,
					MapDepth: 1,
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
						FD:       fd[*protoconf.ActivityConf_Activity]("chapter_map"),
						Depth:    2,
						MapDepth: 2,
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
							FD:       fd[*protoconf.ActivityConf_Activity_Chapter]("section_map"),
							Depth:    3,
							MapDepth: 3,
							NextLevel: &LevelMessage{
								FD:       fd[*protoconf.Section]("section_item_list"),
								Depth:    4,
								MapDepth: 3,
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
									FD:       fd[*protoconf.Section_SectionItem]("decompose_item_list"),
									Depth:    5,
									MapDepth: 3,
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
					FD:       fd[*protoconf.TaskConf]("task_map"),
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
		{
			// Fruit3Conf: list -> list -> map -> list
			// indexes at Country level (L2) and Attr level (L4),
			// ordered index at Item level (L3).
			name: "Fruit3Conf",
			args: args{
				md: md[*protoconf.Fruit3Conf](),
			},
			want: &IndexDescriptor{
				LevelMessage: &LevelMessage{
					FD:       fd[*protoconf.Fruit3Conf]("fruit_list"),
					Depth:    1,
					MapDepth: 0,
					NextLevel: &LevelMessage{
						FD:       fd[*protoconf.Fruit3Conf_Fruit]("country_list"),
						Depth:    2,
						MapDepth: 0,
						Indexes: []*LevelIndex{
							{
								Index: &Index{
									Cols: []string{"CountryName"},
									Name: "",
								},
								MD: md[*protoconf.Fruit3Conf_Fruit_Country](),
								ColFields: []*LevelField{
									{
										FD: fd[*protoconf.Fruit3Conf_Fruit_Country]("name"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											fd[*protoconf.Fruit3Conf_Fruit_Country]("name"),
										},
									},
								},
							},
						},
						NextLevel: &LevelMessage{
							FD:       fd[*protoconf.Fruit3Conf_Fruit_Country]("item_map"),
							Depth:    3,
							MapDepth: 1,
							OrderedIndexes: []*LevelIndex{
								{
									Index: &Index{
										Cols:       []string{"CountryItemPrice"},
										SortedCols: []string{"CountryItemID"},
										Name:       "",
									},
									MD: md[*protoconf.Fruit3Conf_Fruit_Country_Item](),
									ColFields: []*LevelField{
										{
											FD: fd[*protoconf.Fruit3Conf_Fruit_Country_Item]("price"),
											LeveledFDList: []protoreflect.FieldDescriptor{
												fd[*protoconf.Fruit3Conf_Fruit_Country_Item]("price"),
											},
										},
									},
									SortedColFields: []*LevelField{
										{
											FD: fd[*protoconf.Fruit3Conf_Fruit_Country_Item]("id"),
											LeveledFDList: []protoreflect.FieldDescriptor{
												fd[*protoconf.Fruit3Conf_Fruit_Country_Item]("id"),
											},
										},
									},
								},
							},
							NextLevel: &LevelMessage{
								FD:       fd[*protoconf.Fruit3Conf_Fruit_Country_Item]("attr_list"),
								Depth:    4,
								MapDepth: 1,
								Indexes: []*LevelIndex{
									{
										Index: &Index{
											Cols: []string{"CountryItemAttrName"},
											Name: "",
										},
										MD: md[*protoconf.Fruit3Conf_Fruit_Country_Item_Attr](),
										ColFields: []*LevelField{
											{
												FD: fd[*protoconf.Fruit3Conf_Fruit_Country_Item_Attr]("name"),
												LeveledFDList: []protoreflect.FieldDescriptor{
													fd[*protoconf.Fruit3Conf_Fruit_Country_Item_Attr]("name"),
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
		},
		{
			// Fruit5Conf: 3-level map (fruit_map -> country_map -> item_map),
			// but indexes only at MapDepth=2 (Country level).
			// The 3rd level map (MapDepth=3, Item) has no index.
			// This validates that initLevelMessage only collects map keys
			// for levels whose deeper levels have indexes, so len(keys) == 2
			// (not 3), preventing generation of extra LevelIndex key structs.
			name: "Fruit5Conf",
			args: args{
				md: md[*protoconf.Fruit5Conf](),
			},
			want: &IndexDescriptor{
				LevelMessage: &LevelMessage{
					FD:       fd[*protoconf.Fruit5Conf]("fruit_map"),
					Depth:    1,
					MapDepth: 1,
					NextLevel: &LevelMessage{
						FD:       fd[*protoconf.Fruit5Conf_Fruit]("country_map"),
						Depth:    2,
						MapDepth: 2,
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
							FD:       fd[*protoconf.Fruit5Conf_Fruit_Country]("item_map"),
							Depth:    3,
							MapDepth: 3,
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
