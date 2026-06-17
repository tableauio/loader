// Hand-written (custom) messager.
//
// It mirrors the canonical custom-messager pattern of the other-language
// loaders, demonstrating the cross-messager `processAfterLoadAll` extension
// point:
//   - Go:  test/go-tableau-loader/customconf/custom_item_conf.go
//   - C#:  test/csharp-tableau-loader/custom/CustomItemConf.cs
//   - C++: test/cpp-tableau-loader/src/hub/custom/item/custom_item_conf.{h,cpp}
//
// Unlike generated messagers, it carries no config file of its own; its `load`
// is a no-op and it derives its data from another messager (ItemConf) once all
// messagers have finished loading.
import { Messager } from "../tableau/messager.pc.js";
import { Format, type MessagerOptions } from "../tableau/load.pc.js";
import type { Hub } from "../tableau/hub.pc.js";
import * as protoconf from "../tableau/barrel/protoconf.pc.js";

export const CustomItemConfName = "CustomItemConf";

/**
 * CustomItemConf is a hand-written messager. It has no backing file and instead
 * resolves a "special" item from ItemConf during processAfterLoadAll.
 */
export class CustomItemConf extends Messager {
  #specialItemConf: protoconf.ItemConf_Item | undefined;

  /** name returns the CustomItemConf's message name. */
  name(): string {
    return CustomItemConfName;
  }

  /** load is a no-op: this messager has no file of its own. */
  load(_dir: string, _fmt: Format, _options?: MessagerOptions): void {}

  /** processAfterLoadAll consumes ItemConf's data after all messagers load. */
  override processAfterLoadAll(hub: Hub): void {
    const itemConf = hub.getItemConf();
    if (!itemConf) {
      throw new Error("hub get ItemConf failed");
    }
    const conf = itemConf.get1(1);
    if (!conf) {
      throw new Error("hub get item 1 failed");
    }
    this.#specialItemConf = conf;
  }

  /** getSpecialItemName returns the resolved special item's name. */
  getSpecialItemName(): string {
    return this.#specialItemConf?.name ?? "";
  }
}
