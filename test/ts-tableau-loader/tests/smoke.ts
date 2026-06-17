// Smoke-test entry point. Runs the four test blocks in order:
//   Load -> Get -> OrderedMap -> Index
// Each block lives in its own file and shares the harness in ./harness.ts.
import { run as runLoad } from "./load.test.js";
import { run as runGet } from "./get.test.js";
import { run as runOrderedMap } from "./ordered_map.test.js";
import { run as runIndex } from "./index.test.js";
import { passedCount } from "./harness.js";

console.log("# Load");
runLoad();
console.log("# Get");
runGet();
console.log("# OrderedMap");
runOrderedMap();
console.log("# Index");
runIndex();

console.log(`\nAll ${passedCount()} smoke checks passed.`);
