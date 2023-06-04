
var fs = require('fs');
import * as global from "./protoconf/protoconf";
// see properties in protoconf
// console.log(global)
import {protoconf} from "./protoconf/protoconf";
import {tableau} from "./loader";
// // read JSON file and parse ItemConf
let obj = JSON.parse(fs.readFileSync('../../test/testdata/ThemeConf.json', 'utf8'));
// console.log("Hell World!")
// console.log(obj)
let themeConf = protoconf["ThemeConf"].fromObject(obj)
// or let itemConf = protoconf.ThemeConf.fromObject(obj)
console.log(themeConf)

tableau.init()
const hub = new tableau.Hub()
hub.load("../../test/testdata/", tableau.Format.JSON)
console.log(hub.getThemeConf()?.data)

