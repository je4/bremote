(function(t){function e(e){for(var n,s,l=e[0],c=e[1],i=e[2],d=0,m=[];d<l.length;d++)s=l[d],Object.prototype.hasOwnProperty.call(o,s)&&o[s]&&m.push(o[s][0]),o[s]=0;for(n in c)Object.prototype.hasOwnProperty.call(c,n)&&(t[n]=c[n]);u&&u(e);while(m.length)m.shift()();return r.push.apply(r,i||[]),a()}function a(){for(var t,e=0;e<r.length;e++){for(var a=r[e],n=!0,l=1;l<a.length;l++){var c=a[l];0!==o[c]&&(n=!1)}n&&(r.splice(e--,1),t=s(s.s=a[0]))}return t}var n={},o={app:0},r=[];function s(e){if(n[e])return n[e].exports;var a=n[e]={i:e,l:!1,exports:{}};return t[e].call(a.exports,a,a.exports,s),a.l=!0,a.exports}s.m=t,s.c=n,s.d=function(t,e,a){s.o(t,e)||Object.defineProperty(t,e,{enumerable:!0,get:a})},s.r=function(t){"undefined"!==typeof Symbol&&Symbol.toStringTag&&Object.defineProperty(t,Symbol.toStringTag,{value:"Module"}),Object.defineProperty(t,"__esModule",{value:!0})},s.t=function(t,e){if(1&e&&(t=s(t)),8&e)return t;if(4&e&&"object"===typeof t&&t&&t.__esModule)return t;var a=Object.create(null);if(s.r(a),Object.defineProperty(a,"default",{enumerable:!0,value:t}),2&e&&"string"!=typeof t)for(var n in t)s.d(a,n,function(e){return t[e]}.bind(null,n));return a},s.n=function(t){var e=t&&t.__esModule?function(){return t["default"]}:function(){return t};return s.d(e,"a",e),e},s.o=function(t,e){return Object.prototype.hasOwnProperty.call(t,e)},s.p="/";var l=window["webpackJsonp"]=window["webpackJsonp"]||[],c=l.push.bind(l);l.push=e,l=l.slice();for(var i=0;i<l.length;i++)e(l[i]);var u=c;r.push([0,"chunk-vendors"]),a()})({0:function(t,e,a){t.exports=a("56d7")},1229:function(t,e,a){"use strict";var n=a("8704"),o=a.n(n);o.a},5114:function(t,e,a){},"56d7":function(t,e,a){"use strict";a.r(e);a("2ca0"),a("e260"),a("e6cf"),a("cca6"),a("a79d");var n=a("2b0e"),o=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("v-app",[a("BRemoteMain")],1)},r=[],s=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",[a("nav",[a("v-app-bar",{attrs:{app:""}},[a("v-app-bar-nav-icon",{staticClass:"hidden-md-and-up",on:{click:function(e){e.stopPropagation(),t.drawer=!t.drawer}}}),a("v-toolbar-title",{staticClass:"headline text-uppercase"},[a("span",[t._v("RemoteScreen")]),a("span",{staticClass:"font-weight-light"},[t._v("Controller")])])],1),a("v-navigation-drawer",{attrs:{app:"",left:""},model:{value:t.drawer,callback:function(e){t.drawer=e},expression:"drawer"}},[a("v-toolbar",{attrs:{flat:""}},[a("v-list",[a("v-list-item",[a("v-list-item-title",{staticClass:"title"},[t._v("Screens")])],1)],1)],1),a("v-divider"),a("v-list",t._l(t.orderedClients,(function(e){return a("v-list-item",{key:e.InstanceName,attrs:{exact:""},on:{click:function(a){return t.selectMachine(e.InstanceName)}}},[a("v-list-item-action",[a("v-icon",[t._v(t._s(t.machineIcon))])],1),a("v-list-item-content",[t._v(t._s(e.InstanceName))])],1)})),1)],1)],1),a("v-content",{staticClass:"ma-3"},[a("Machine",{attrs:{name:t.currMachine}})],1)],1)},l=[],c=a("bc3a"),i=a.n(c),u=a("2ef0"),d=a.n(u),m=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",[a("div",[a("h1",{staticClass:"title"},[t._v(t._s(t.name))])]),""!=t.name?a("div",[a("v-tabs",{attrs:{"background-color":"transparent",color:"basil",grow:""},model:{value:t.tab,callback:function(e){t.tab=e},expression:"tab"}},[a("v-tab",{key:"screenshot"},[t._v(" Screenshot ")]),a("v-tab",{key:"console"},[t._v(" Browser Console ")]),a("v-tab",{key:"url"},[t._v(" Send URL ")]),a("v-tab",{key:"command"},[t._v(" Send Command ")])],1),a("v-tabs-items",{model:{value:t.tab,callback:function(e){t.tab=e},expression:"tab"}},[a("v-tab-item",{key:"screenshot",staticClass:"ma-3"},[a("Screenshot",{attrs:{name:t.name}})],1),a("v-tab-item",{key:"console",staticClass:"ma-3"},[a("Console",{attrs:{name:t.name}})],1),a("v-tab-item",{key:"url",staticClass:"ma-3"},[a("UploadURL",{attrs:{name:t.name}})],1),a("v-tab-item",{key:"command",staticClass:"ma-3"},[a("Template",{attrs:{name:t.name}})],1)],1)],1):t._e()])},v=[],p=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("v-container",{staticClass:"grey lighten-5"},[a("v-row",{staticClass:"mb-6",attrs:{"no-gutters":""}},[a("v-col",[a("div",{staticClass:"screenshot"},[a("img",{staticStyle:{"max-width":"100%","max-height":"100%"},attrs:{src:t.screenshot}})])]),a("v-col",[a("v-btn",{staticClass:"mx-2",attrs:{fab:"",dark:"",color:"blue"},on:{click:t.reloadScreenshot}},[a("v-icon",{attrs:{dark:""}},[t._v(t._s(t.reloadIcon))])],1)],1)],1)],1)},f=[],h=(a("0d03"),a("b0c0"),a("94ed")),b={name:"Screenshot",props:{name:{type:String}},data:function(){return{screenshot:null,reloadIcon:h["b"]}},watch:{name:function(){this.reloadScreenshot()}},methods:{reloadScreenshot:function(){this.screenshot="";var t=this,e=new Image;e.onload=function(){t.screenshot=this.src},e.src=this.$controllerBase+"/"+this.name+"/screenshot/medium?"+(new Date).getTime()}},mounted:function(){this.reloadScreenshot()}},g=b,y=(a("1229"),a("2877")),w=a("6544"),C=a.n(w),V=a("8336"),_=a("62ad"),x=a("a523"),k=a("132d"),I=a("0fd9"),S=Object(y["a"])(g,p,f,!1,null,"6a34aab7",null),T=S.exports;C()(S,{VBtn:V["a"],VCol:_["a"],VContainer:x["a"],VIcon:k["a"],VRow:I["a"]});var B=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("v-container",{staticClass:"grey lighten-5"},[a("v-row",{staticClass:"mb-6",attrs:{"no-gutters":""}},[a("v-col",[a("div",{staticClass:"screenshot scroll"},[a("vue-code-highlight",{staticStyle:{"min-width":"100%"},attrs:{fluid:""}},[t._v(t._s(t.console))])],1)]),a("v-col",[a("v-btn",{staticClass:"mx-2",attrs:{fab:"",dark:"",color:"blue"},on:{click:t.reloadConsole}},[a("v-icon",{attrs:{dark:""}},[t._v(t._s(t.reloadIcon))])],1)],1)],1)],1)},O=[],j=(a("a4d3"),a("e01a"),a("d28b"),a("26e9"),a("d3b7"),a("3ca3"),a("ddb0"),a("d36c")),R={name:"Console",components:{VueCodeHighlight:j["a"]},props:{name:{type:String}},data:function(){return{console:"",browserlog:[],reloadIcon:h["b"]}},watch:{name:function(){this.reloadConsole()}},methods:{reloadConsole:function(){var t=this;this.console="",i.a.get(this.$controllerBase+"/client/"+this.name+"/browserlog").then((function(e){if(t.console="",200!=e.status&&alert("browserlog: "+e.statusText),null!=e.data){var a=!0,n=!1,o=void 0;try{for(var r,s=e.data.reverse()[Symbol.iterator]();!(a=(r=s.next()).done);a=!0){var l=r.value;t.console+=l+"\n"}}catch(c){n=!0,o=c}finally{try{a||null==s.return||s.return()}finally{if(n)throw o}}}}))}},mounted:function(){this.reloadConsole()}},$=R,M=(a("81c6"),Object(y["a"])($,B,O,!1,null,"515faf9f",null)),L=M.exports;C()(M,{VBtn:V["a"],VCol:_["a"],VContainer:x["a"],VIcon:k["a"],VRow:I["a"]});var U=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("v-form",[a("v-container",[a("v-row",[a("v-col",{attrs:{cols:"24",sm:"12",md:"6"}},[a("v-text-field",{attrs:{label:"URL",placeholder:"https://"},model:{value:t.target,callback:function(e){t.target=e},expression:"target"}})],1),a("v-col",{attrs:{cols:"24",sm:"6",md:"3"}},[a("v-btn",{staticClass:"mx-2",attrs:{fab:"",dark:"",color:"blue"},on:{click:t.sendToDisplay}},[a("v-icon",{attrs:{dark:""}},[t._v(t._s(t.uploadIcon))])],1)],1)],1)],1)],1)},P=[],E=(a("caad"),a("2532"),{name:"UploadURL",props:{name:{type:String}},components:{},data:function(){return{target:"",uploadIcon:h["c"],tab:null}},watch:{},methods:{sendToDisplay:function(){this.target.includes("://")&&i.a.post(this.$controllerBase+"/client/"+this.name+"/navigate",{url:this.target,nextstatus:"testing"}).then()}}}),J=E,N=(a("7fe8"),a("4bd4")),D=a("8654"),A=Object(y["a"])(J,U,P,!1,null,"359ebd92",null),F=A.exports;C()(A,{VBtn:V["a"],VCol:_["a"],VContainer:x["a"],VForm:N["a"],VIcon:k["a"],VRow:I["a"],VTextField:D["a"]});var H=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("v-form",[a("v-container",[a("v-row",[a("v-col",{attrs:{cols:"24",sm:"12",md:"6"}},[a("v-select",{attrs:{items:t.controllers,"item-text":"InstanceName",label:"Controller",dense:"",solo:""},model:{value:t.controller,callback:function(e){t.controller=e},expression:"controller"}}),a("v-select",{attrs:{items:t.templates,label:"Template",dense:"",solo:""},model:{value:t.template,callback:function(e){t.template=e},expression:"template"}}),a("v-text-field",{attrs:{label:"URL",placeholder:"https://"},model:{value:t.target,callback:function(e){t.target=e},expression:"target"}}),a("br"),a("v-jsoneditor",{attrs:{plus:!1,height:"400px"},model:{value:t.targetJson,callback:function(e){t.targetJson=e},expression:"targetJson"}})],1),a("v-col",{attrs:{cols:"24",sm:"6",md:"3"}},[a("v-btn",{staticClass:"mx-2",attrs:{fab:"",dark:"",color:"blue"},on:{click:t.sendToDisplay}},[a("v-icon",{attrs:{dark:""}},[t._v(t._s(t.uploadIcon))])],1)],1)],1)],1)],1)},W=[],q=a("27f7"),z={name:"UploadURL",props:{name:{type:String}},components:{VJsoneditor:q["a"]},data:function(){return{target:"",uploadIcon:h["c"],tab:null,controllers:[],controller:"",templates:[],template:""}},watch:{controller:function(){var t=this;i.a.get(this.$controllerBase+"/controller/"+this.controller+"/templates").then((function(e){t.templates=e.data}))}},mounted:function(){var t=this;i.a.get(this.$controllerBase+"/controller").then((function(e){t.controllers=e.data}))},methods:{}},G=z,K=a("b974"),Q=Object(y["a"])(G,H,W,!1,null,"2a921dc3",null),X=Q.exports;C()(Q,{VBtn:V["a"],VCol:_["a"],VContainer:x["a"],VForm:N["a"],VIcon:k["a"],VRow:I["a"],VSelect:K["a"],VTextField:D["a"]});var Y={name:"Machine",props:{name:{type:String}},components:{UploadURL:F,Screenshot:T,Console:L,Template:X},data:function(){return{target:"",targetJson:{},tab:null}},watch:{},methods:{}},Z=Y,tt=(a("d4c7"),a("71a3")),et=a("c671"),at=a("fe57"),nt=a("aac8"),ot=Object(y["a"])(Z,m,v,!1,null,"4ed5ad7c",null),rt=ot.exports;C()(ot,{VTab:tt["a"],VTabItem:et["a"],VTabs:at["a"],VTabsItems:nt["a"]});var st={name:"BRemoteMain",components:{Machine:rt},data:function(){return{drawer:!0,machineIcon:h["a"],clients:[],currMachine:""}},computed:{orderedClients:function(){return d.a.orderBy(this.clients,"InstanceName")}},methods:{selectMachine:function(t){this.currMachine=t}},mounted:function(){var t=this;i.a.get(this.$controllerBase+"/client").then((function(e){return t.clients=e.data}))}},lt=st,ct=a("40dc"),it=a("5bc1"),ut=a("a75b"),dt=a("ce7e"),mt=a("8860"),vt=a("da13"),pt=a("1800"),ft=a("5d23"),ht=a("f774"),bt=a("71d9"),gt=a("2a7f"),yt=Object(y["a"])(lt,s,l,!1,null,null,null),wt=yt.exports;C()(yt,{VAppBar:ct["a"],VAppBarNavIcon:it["a"],VContent:ut["a"],VDivider:dt["a"],VIcon:k["a"],VList:mt["a"],VListItem:vt["a"],VListItemAction:pt["a"],VListItemContent:ft["a"],VListItemTitle:ft["b"],VNavigationDrawer:ht["a"],VToolbar:bt["a"],VToolbarTitle:gt["a"]});var Ct={name:"App",components:{BRemoteMain:wt},data:function(){return{drawer:!0}}},Vt=Ct,_t=a("7496"),xt=Object(y["a"])(Vt,o,r,!1,null,null,null),kt=xt.exports;C()(xt,{VApp:_t["a"]});var It=a("f309");n["a"].use(It["a"]);var St=new It["a"]({icons:{iconfont:"mdiSvg"}}),Tt=a("dc60");n["a"].config.productionTip=!1;var Bt="";try{var Ot=i.a.get(""),jt=Ot.headers.get;jt["server"].startsWith("RemoteScreenController")||(Bt="https://127.0.0.1:8444",n["a"].use(Tt["a"],{globals:{$controllerBase:Bt}}))}catch(Rt){Bt="https://127.0.0.1:8444",n["a"].use(Tt["a"],{globals:{$controllerBase:Bt}})}new n["a"]({vuetify:St,render:function(t){return t(kt)}}).$mount("#app")},"7fe8":function(t,e,a){"use strict";var n=a("fb38"),o=a.n(n);o.a},"81c6":function(t,e,a){"use strict";var n=a("eb22"),o=a.n(n);o.a},8704:function(t,e,a){},d4c7:function(t,e,a){"use strict";var n=a("5114"),o=a.n(n);o.a},eb22:function(t,e,a){},fb38:function(t,e,a){}});
//# sourceMappingURL=app.33668056.js.map