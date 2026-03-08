// Copyright 2026 chenyang
// 
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// 
//     https://www.apache.org/licenses/LICENSE-2.0
// 
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import fs from "fs"


const file = "frontend/package.json"
const part = process.env.PART || "patch"
const maxRemote = process.env.MAX_REMOTE_VERSION || "0.0.0"

const pkg = JSON.parse(fs.readFileSync(file, "utf8"))

function parse(v) {
  return (v || "0.0.0")
    .split(".")
    .map(n => parseInt(n, 10) || 0)
    .concat([0,0,0])
    .slice(0,3)
}

function cmp(a,b){
  return a[0]-b[0] || a[1]-b[1] || a[2]-b[2]
}

let seg = cmp(parse(pkg.version), parse(maxRemote)) >= 0
  ? parse(pkg.version)
  : parse(maxRemote)

if(part==="major"){
  seg[0]++
  seg[1]=0
  seg[2]=0
}else if(part==="minor"){
  seg[1]++
  seg[2]=0
}else{
  seg[2]++
}

const next = seg.join(".")

pkg.version = next

fs.writeFileSync(file, JSON.stringify(pkg,null,2)+"\n")

process.stdout.write(next)