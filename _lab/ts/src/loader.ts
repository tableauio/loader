const fs = require('fs');
const path = require('path');

import { error } from "console";
import { protoconf } from "./protoconf/protoconf"
import { fstat } from "fs";
export namespace tableau {
    export enum Format {
        JSON,
        BIN,
        TXT,
    }
    export enum Code {
        SUCCESS,
        NOT_FOUND,
        NULL,
        ILLEGAL_PARAM,
    }
    export interface Error {
        code: Code;
        msg?: string;
    }

    interface Messager {
        name: string
        data: any
        load(dir: string, fmt: Format): Error | null;
        create(): Messager
    }

    type MessagerGenerator = () => Messager

    class Registrar {
        private static instance: Registrar;
        public generators: Map<string, MessagerGenerator> = new Map<string, MessagerGenerator>();
        private constructor() { }
        public static get Instance(): Registrar {
            return this.instance || (this.instance = new this());
        }
        public register(messager: Messager) {
            this.generators.set(messager.name, messager.create)
        }
    }

    export function init() {
        Registrar.Instance.register(new ThemeConf())
    }

    type MessagerMap = Map<string, Messager>

    export class Hub {
        private messagerMap: MessagerMap;
        public constructor() {
            this.messagerMap = new Map<string, Messager>();
        }

        public setMessagerMap(messagerMap: MessagerMap): void {
            this.messagerMap = messagerMap
        }
        public newMessagerMap(): MessagerMap {
            let messagerMap = new Map<string, Messager>()
            Registrar.Instance.generators.forEach((value: MessagerGenerator, key: string) => {
                console.log(key);
                messagerMap.set(key, value())
            });
            return messagerMap
        }
        public load(dir: string, fmt: Format): Error | null {
            let messagerMap = this.newMessagerMap()
            messagerMap.forEach((msger: Messager, name: string) => {
                let err = msger.load(dir, fmt)
                if (err) {
                    return err
                }
                console.log("Loaded: " + name);
            });
            this.setMessagerMap(messagerMap)
            return null;
        }
        public getThemeConf(): ThemeConf | null {
            const msger = this.messagerMap.get("ThemeConf")
            if (msger) {
                return msger as unknown as ThemeConf
            }
            return null
        }
    }

    function load(messager: Messager, dir: string, fmt: Format): Error | null {
        switch (fmt) {
            case Format.JSON: {
                const fpath = path.join(dir, messager.name + ".json")
                let obj = JSON.parse(fs.readFileSync(fpath, 'utf8'));
                messager.data = (protoconf as any)[messager.name].fromObject(obj)
                return null
            }
            case Format.TXT: {
                return { code: Code.ILLEGAL_PARAM, msg: "not supported yet" }
            }
            case Format.BIN: {
                let fpath = path.join(dir, messager.name + ".bin")
                let buffer = fs.readFileSync(fpath, null);
                messager.data = (protoconf as any)[messager.name].decode(buffer)
                return null
            }
            default:
                return { code: Code.ILLEGAL_PARAM, msg: "unknown format" }
        }
    }

    export class ThemeConf {
        public name: string
        public data: protoconf.ThemeConf
        public constructor() {
            this.name = "ThemeConf"
            this.data = new protoconf.ThemeConf()
        }
        public load(dir: string, fmt: Format): Error | null {
            console.log("loading...")
            return load(this, dir, fmt)
        }
        public create(): Messager {
            return new ThemeConf()
        }
    }
}