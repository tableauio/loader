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
					FD: fd[*protoconf.ItemConf]("item_map"),
					NextLevel: &LevelMessage{
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
									Cols: []string{"Param"},
									Keys: []string{"ID"},
									Name: "ItemInfo",
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
								KeyFields: []*LevelField{
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
									Cols: []string{"ID", "Name"},
									Keys: []string{"Type", "UseEffectType"},
									Name: "AwardItem",
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
								KeyFields: []*LevelField{
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
					FD: fd[*protoconf.HeroConf]("hero_map"),
					NextLevel: &LevelMessage{
						FD: fd[*protoconf.HeroConf_Hero]("attr_map"),
						NextLevel: &LevelMessage{
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
					FD: fd[*protoconf.ActivityConf]("activity_map"),
					NextLevel: &LevelMessage{
						FD: fd[*protoconf.ActivityConf_Activity]("chapter_map"),
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
							FD: fd[*protoconf.ActivityConf_Activity_Chapter]("section_map"),
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
										Cols: []string{"ChapterName"},
										Keys: []string{"AwardID"},
										Name: "NamedChapter",
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
									KeyFields: []*LevelField{
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
								FD: fd[*protoconf.Section]("section_item_list"),
								NextLevel: &LevelMessage{
									FD: fd[*protoconf.Section_SectionItem]("decompose_item_list"),
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
									NextLevel: &LevelMessage{},
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
					FD: fd[*protoconf.TaskConf]("task_map"),
					NextLevel: &LevelMessage{
						Indexes: []*LevelIndex{
							{
								Index: &Index{
									Cols: []string{"ActivityID"},
									Keys: []string{"Goal", "ID"},
									Name: "",
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
								KeyFields: []*LevelField{
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
