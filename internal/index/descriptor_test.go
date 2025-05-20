package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tableauio/loader/test/go-tableau-loader/protoconf"
	"google.golang.org/protobuf/reflect/protoreflect"
)

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
				md: (&protoconf.ItemConf{}).ProtoReflect().Descriptor(),
			},
			want: &IndexDescriptor{
				LevelMessage: &LevelMessage{
					FD: (&protoconf.ItemConf{}).ProtoReflect().Descriptor().Fields().ByName("item_map"),
					NextLevel: &LevelMessage{
						Indexes: []*LevelIndex{
							{
								Index: &Index{
									Cols: []string{"Type"},
									Name: "",
								},
								MD: (&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor(),
								ColFields: []*LevelField{
									{
										FD: (&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("type"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											(&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("type"),
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
								MD: (&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor(),
								ColFields: []*LevelField{
									{
										FD: (&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("param_list"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											(&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("param_list"),
										},
									},
								},
								KeyFields: []*LevelField{
									{
										FD: (&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("id"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											(&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("id"),
										},
									},
								},
							},
							{
								Index: &Index{
									Cols: []string{"Default"},
									Name: "ItemDefaultInfo",
								},
								MD: (&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor(),
								ColFields: []*LevelField{
									{
										FD: (&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("default"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											(&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("default"),
										},
									},
								},
							},
							{
								Index: &Index{
									Cols: []string{"ExtType"},
									Name: "ItemExtInfo",
								},
								MD: (&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor(),
								ColFields: []*LevelField{
									{
										FD: (&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("ext_type_list"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											(&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("ext_type_list"),
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
								MD: (&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor(),
								ColFields: []*LevelField{
									{
										FD: (&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("id"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											(&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("id"),
										},
									},
									{
										FD: (&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("name"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											(&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("name"),
										},
									},
								},
								KeyFields: []*LevelField{
									{
										FD: (&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("type"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											(&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("type"),
										},
									},
									{
										FD: (&protoconf.UseEffect{}).ProtoReflect().Descriptor().Fields().ByName("type"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											(&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("use_effect"),
											(&protoconf.UseEffect{}).ProtoReflect().Descriptor().Fields().ByName("type"),
										},
									},
								},
							},
							{
								Index: &Index{
									Cols: []string{"ID", "Type", "Param", "ExtType"},
									Name: "SpecialItem",
								},
								MD: (&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor(),
								ColFields: []*LevelField{
									{
										FD: (&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("id"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											(&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("id"),
										},
									},
									{
										FD: (&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("type"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											(&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("type"),
										},
									},
									{
										FD: (&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("param_list"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											(&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("param_list"),
										},
									},
									{
										FD: (&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("ext_type_list"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											(&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("ext_type_list"),
										},
									},
								},
							},
							{
								Index: &Index{
									Cols: []string{"PathDir"},
									Name: "ItemPathDir",
								},
								MD: (&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor(),
								ColFields: []*LevelField{
									{
										FD: (&protoconf.Path{}).ProtoReflect().Descriptor().Fields().ByName("dir"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											(&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("path"),
											(&protoconf.Path{}).ProtoReflect().Descriptor().Fields().ByName("dir"),
										},
									},
								},
							},
							{
								Index: &Index{
									Cols: []string{"PathName"},
									Name: "ItemPathName",
								},
								MD: (&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor(),
								ColFields: []*LevelField{
									{
										FD: (&protoconf.Path{}).ProtoReflect().Descriptor().Fields().ByName("name_list"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											(&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("path"),
											(&protoconf.Path{}).ProtoReflect().Descriptor().Fields().ByName("name_list"),
										},
									},
								},
							},
							{
								Index: &Index{
									Cols: []string{"PathFriendID"},
									Name: "ItemPathFriendID",
								},
								MD: (&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor(),
								ColFields: []*LevelField{
									{
										FD: (&protoconf.Path_Friend{}).ProtoReflect().Descriptor().Fields().ByName("id"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											(&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("path"),
											(&protoconf.Path{}).ProtoReflect().Descriptor().Fields().ByName("friend"),
											(&protoconf.Path_Friend{}).ProtoReflect().Descriptor().Fields().ByName("id"),
										},
									},
								},
							},
							{
								Index: &Index{
									Cols: []string{"UseEffectType"},
									Name: "UseEffectType",
								},
								MD: (&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor(),
								ColFields: []*LevelField{
									{
										FD: (&protoconf.UseEffect{}).ProtoReflect().Descriptor().Fields().ByName("type"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											(&protoconf.ItemConf_Item{}).ProtoReflect().Descriptor().Fields().ByName("use_effect"),
											(&protoconf.UseEffect{}).ProtoReflect().Descriptor().Fields().ByName("type"),
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
				md: (&protoconf.HeroConf{}).ProtoReflect().Descriptor(),
			},
			want: &IndexDescriptor{
				LevelMessage: &LevelMessage{
					FD: (&protoconf.HeroConf{}).ProtoReflect().Descriptor().Fields().ByName("hero_map"),
					NextLevel: &LevelMessage{
						FD: (&protoconf.HeroConf_Hero{}).ProtoReflect().Descriptor().Fields().ByName("attr_map"),
						NextLevel: &LevelMessage{
							Indexes: []*LevelIndex{
								{
									Index: &Index{
										Cols: []string{"Title"},
										Name: "",
									},
									MD: (&protoconf.HeroConf_Hero_Attr{}).ProtoReflect().Descriptor(),
									ColFields: []*LevelField{
										{
											FD: (&protoconf.HeroConf_Hero_Attr{}).ProtoReflect().Descriptor().Fields().ByName("title"),
											LeveledFDList: []protoreflect.FieldDescriptor{
												(&protoconf.HeroConf_Hero_Attr{}).ProtoReflect().Descriptor().Fields().ByName("title"),
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
				md: (&protoconf.ActivityConf{}).ProtoReflect().Descriptor(),
			},
			want: &IndexDescriptor{
				LevelMessage: &LevelMessage{
					FD: (&protoconf.ActivityConf{}).ProtoReflect().Descriptor().Fields().ByName("activity_map"),
					NextLevel: &LevelMessage{
						FD: (&protoconf.ActivityConf_Activity{}).ProtoReflect().Descriptor().Fields().ByName("chapter_map"),
						Indexes: []*LevelIndex{
							{
								Index: &Index{
									Cols: []string{"ActivityName"},
									Name: "",
								},
								MD: (&protoconf.ActivityConf_Activity{}).ProtoReflect().Descriptor(),
								ColFields: []*LevelField{
									{
										FD: (&protoconf.ActivityConf_Activity{}).ProtoReflect().Descriptor().Fields().ByName("activity_name"),
										LeveledFDList: []protoreflect.FieldDescriptor{
											(&protoconf.ActivityConf_Activity{}).ProtoReflect().Descriptor().Fields().ByName("activity_name"),
										},
									},
								},
							},
						},
						NextLevel: &LevelMessage{
							FD: (&protoconf.ActivityConf_Activity_Chapter{}).ProtoReflect().Descriptor().Fields().ByName("section_map"),
							Indexes: []*LevelIndex{
								{
									Index: &Index{
										Cols: []string{"ChapterID"},
										Name: "",
									},
									MD: (&protoconf.ActivityConf_Activity_Chapter{}).ProtoReflect().Descriptor(),
									ColFields: []*LevelField{
										{
											FD: (&protoconf.ActivityConf_Activity_Chapter{}).ProtoReflect().Descriptor().Fields().ByName("chapter_id"),
											LeveledFDList: []protoreflect.FieldDescriptor{
												(&protoconf.ActivityConf_Activity_Chapter{}).ProtoReflect().Descriptor().Fields().ByName("chapter_id"),
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
									MD: (&protoconf.ActivityConf_Activity_Chapter{}).ProtoReflect().Descriptor(),
									ColFields: []*LevelField{
										{
											FD: (&protoconf.ActivityConf_Activity_Chapter{}).ProtoReflect().Descriptor().Fields().ByName("chapter_name"),
											LeveledFDList: []protoreflect.FieldDescriptor{
												(&protoconf.ActivityConf_Activity_Chapter{}).ProtoReflect().Descriptor().Fields().ByName("chapter_name"),
											},
										},
									},
									KeyFields: []*LevelField{
										{
											FD: (&protoconf.ActivityConf_Activity_Chapter{}).ProtoReflect().Descriptor().Fields().ByName("award_id"),
											LeveledFDList: []protoreflect.FieldDescriptor{
												(&protoconf.ActivityConf_Activity_Chapter{}).ProtoReflect().Descriptor().Fields().ByName("award_id"),
											},
										},
									},
								},
							},
							NextLevel: &LevelMessage{
								FD: (&protoconf.Section{}).ProtoReflect().Descriptor().Fields().ByName("section_item_list"),
								NextLevel: &LevelMessage{
									Indexes: []*LevelIndex{
										{
											Index: &Index{
												Cols: []string{"SectionItemID"},
												Name: "Award",
											},
											MD: (&protoconf.Item{}).ProtoReflect().Descriptor(),
											ColFields: []*LevelField{
												{
													FD: (&protoconf.Item{}).ProtoReflect().Descriptor().Fields().ByName("id"),
													LeveledFDList: []protoreflect.FieldDescriptor{
														(&protoconf.Item{}).ProtoReflect().Descriptor().Fields().ByName("id"),
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseIndexDescriptor(tt.args.md)
			assert.EqualValues(t, tt.want, got)
		})
	}
}
