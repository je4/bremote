(function(t){function a(a){for(var n,s,c=a[0],l=a[1],i=a[2],d=0,v=[];d<c.length;d++)s=c[d],Object.prototype.hasOwnProperty.call(o,s)&&o[s]&&v.push(o[s][0]),o[s]=0;for(n in l)Object.prototype.hasOwnProperty.call(l,n)&&(t[n]=l[n]);u&&u(a);while(v.length)v.shift()();return r.push.apply(r,i||[]),e()}function e(){for(var t,a=0;a<r.length;a++){for(var e=r[a],n=!0,c=1;c<e.length;c++){var l=e[c];0!==o[l]&&(n=!1)}n&&(r.splice(a--,1),t=s(s.s=e[0]))}return t}var n={},o={app:0},r=[];function s(a){if(n[a])return n[a].exports;var e=n[a]={i:a,l:!1,exports:{}};return t[a].call(e.exports,e,e.exports,s),e.l=!0,e.exports}s.m=t,s.c=n,s.d=function(t,a,e){s.o(t,a)||Object.defineProperty(t,a,{enumerable:!0,get:e})},s.r=function(t){"undefined"!==typeof Symbol&&Symbol.toStringTag&&Object.defineProperty(t,Symbol.toStringTag,{value:"Module"}),Object.defineProperty(t,"__esModule",{value:!0})},s.t=function(t,a){if(1&a&&(t=s(t)),8&a)return t;if(4&a&&"object"===typeof t&&t&&t.__esModule)return t;var e=Object.create(null);if(s.r(e),Object.defineProperty(e,"default",{enumerable:!0,value:t}),2&a&&"string"!=typeof t)for(var n in t)s.d(e,n,function(a){return t[a]}.bind(null,n));return e},s.n=function(t){var a=t&&t.__esModule?function(){return t["default"]}:function(){return t};return s.d(a,"a",a),a},s.o=function(t,a){return Object.prototype.hasOwnProperty.call(t,a)},s.p="/static/ui/";var c=window["webpackJsonp"]=window["webpackJsonp"]||[],l=c.push.bind(c);c.push=a,c=c.slice();for(var i=0;i<c.length;i++)a(c[i]);var u=l;r.push([0,"chunk-vendors"]),e()})({0:function(t,a,e){t.exports=e("56d7")},3536:function(t,a,e){"use strict";var n=e("8a3a"),o=e.n(n);o.a},5114:function(t,a,e){},"56d7":function(t,a,e){"use strict";e.r(a);e("e260"),e("e6cf"),e("cca6"),e("a79d");var n=e("2b0e"),o=function(){var t=this,a=t.$createElement,e=t._self._c||a;return e("v-app",[e("BRemoteMain")],1)},r=[],s=function(){var t=this,a=t.$createElement,e=t._self._c||a;return e("div",[e("nav",[e("v-app-bar",{attrs:{app:""}},[e("v-app-bar-nav-icon",{staticClass:"hidden-md-and-up",on:{click:function(a){a.stopPropagation(),t.drawer=!t.drawer}}}),e("v-toolbar-title",{staticClass:"headline text-uppercase"},[e("span",[t._v("RemoteScreen")]),e("span",{staticClass:"font-weight-light"},[t._v("Controller")])])],1),e("v-navigation-drawer",{attrs:{app:"",left:""},model:{value:t.drawer,callback:function(a){t.drawer=a},expression:"drawer"}},[e("v-toolbar",{attrs:{flat:""}},[e("v-list",[e("v-list-item",[e("v-list-item-title",{staticClass:"title"},[t._v("Screens")])],1)],1)],1),e("v-divider"),e("v-list",t._l(t.orderedClients,(function(a){return e("v-list-item",{key:a.InstanceName,attrs:{exact:""},on:{click:function(e){return t.selectMachine(a.InstanceName)}}},[e("v-list-item-action",[e("v-icon",[t._v(t._s(t.machineIcon))])],1),e("v-list-item-content",[t._v(t._s(a.InstanceName))])],1)})),1)],1)],1),e("v-content",{staticClass:"ma-3"},[e("Machine",{attrs:{name:t.currMachine}})],1)],1)},c=[],l=e("bc3a"),i=e.n(l),u=e("2ef0"),d=e.n(u),v=function(){var t=this,a=t.$createElement,e=t._self._c||a;return e("div",[e("div",[e("h1",{staticClass:"title"},[t._v(t._s(t.name))])]),""!=t.name?e("div",[e("v-tabs",{attrs:{"background-color":"transparent",color:"basil",grow:""},model:{value:t.tab,callback:function(a){t.tab=a},expression:"tab"}},[e("v-tab",{key:"screenshot"},[t._v(" Screenshot ")]),e("v-tab",{key:"console"},[t._v(" Browser Console ")]),e("v-tab",{key:"url"},[t._v(" Send URL ")]),e("v-tab",{key:"command"},[t._v(" Send Command ")])],1),e("v-tabs-items",{model:{value:t.tab,callback:function(a){t.tab=a},expression:"tab"}},[e("v-tab-item",{key:"screenshot",staticClass:"ma-3"},[e("Screenshot",{attrs:{name:t.name}})],1),e("v-tab-item",{key:"console",staticClass:"ma-3"},[e("Console",{attrs:{name:t.name}})],1),e("v-tab-item",{key:"url",staticClass:"ma-3"},[e("UploadURL",{attrs:{name:t.name}})],1),e("v-tab-item",{key:"command",staticClass:"ma-3"},[e("Template",{attrs:{name:t.name}})],1)],1)],1):t._e()])},b=[],m=function(){var t=this,a=t.$createElement,e=t._self._c||a;return e("v-container",{staticClass:"lighten-5"},[e("v-row",{staticClass:"mb-6",attrs:{"no-gutters":""}},[e("v-col",[e("div",{staticClass:"screenshot"},[e("img",{staticStyle:{"max-width":"100%","max-height":"100%"},attrs:{src:t.screenshot}})])]),e("v-col",[e("v-btn",{staticClass:"mx-2",attrs:{fab:"",dark:"",color:"blue"},on:{click:t.reloadScreenshot}},[e("v-icon",{attrs:{dark:""}},[t._v(t._s(t.reloadIcon))])],1)],1)],1)],1)},p=[],f=(e("0d03"),e("b0c0"),e("94ed")),h={name:"Screenshot",props:{name:{type:String}},data:function(){return{screenshot:null,reloadIcon:f["b"]}},watch:{name:function(){this.reloadScreenshot()}},methods:{reloadScreenshot:function(){this.screenshot="";var t=this,a=new Image;a.onload=function(){t.screenshot=this.src},a.src=this.$controllerBase+"/"+this.name+"/screenshot/medium?"+(new Date).getTime()}},mounted:function(){this.reloadScreenshot()}},k=h,g=(e("3536"),e("2877")),x=e("6544"),w=e.n(x),y=e("8336"),_=e("62ad"),V=e("a523"),C=e("132d"),S=e("0fd9"),I=Object(g["a"])(k,m,p,!1,null,"78ad0a33",null),B=I.exports;w()(I,{VBtn:y["a"],VCol:_["a"],VContainer:V["a"],VIcon:C["a"],VRow:S["a"]});var O=function(){var t=this,a=t.$createElement,e=t._self._c||a;return e("v-container",{staticClass:"grey lighten-5"},[e("v-row",{staticClass:"mb-6",attrs:{"no-gutters":""}},[e("v-col",[e("div",{staticClass:"screenshot scroll"},[e("vue-code-highlight",{staticStyle:{"min-width":"100%"},attrs:{fluid:""}},[t._v(t._s(t.console))])],1)]),e("v-col",[e("v-btn",{staticClass:"mx-2",attrs:{fab:"",dark:"",color:"blue"},on:{click:t.reloadConsole}},[e("v-icon",{attrs:{dark:""}},[t._v(t._s(t.reloadIcon))])],1)],1)],1)],1)},T=[],$=(e("a4d3"),e("e01a"),e("d28b"),e("26e9"),e("d3b7"),e("3ca3"),e("ddb0"),e("d36c")),N={name:"Console",components:{VueCodeHighlight:$["a"]},props:{name:{type:String}},data:function(){return{console:"",browserlog:[],reloadIcon:f["b"]}},watch:{name:function(){this.reloadConsole()}},methods:{reloadConsole:function(){var t=this;this.console="",i.a.get(this.$controllerBase+"/client/"+this.name+"/browserlog").then((function(a){if(t.console="",200!=a.status&&alert("browserlog: "+a.statusText),null!=a.data){var e=!0,n=!1,o=void 0;try{for(var r,s=a.data.reverse()[Symbol.iterator]();!(e=(r=s.next()).done);e=!0){var c=r.value;t.console+=c+"\n"}}catch(l){n=!0,o=l}finally{try{e||null==s.return||s.return()}finally{if(n)throw o}}}}))}},mounted:function(){this.reloadConsole()}},j=N,M=(e("81c6"),Object(g["a"])(j,O,T,!1,null,"515faf9f",null)),R=M.exports;w()(M,{VBtn:y["a"],VCol:_["a"],VContainer:V["a"],VIcon:C["a"],VRow:S["a"]});var J=function(){var t=this,a=t.$createElement,e=t._self._c||a;return e("v-form",[e("v-container",[e("v-row",[e("v-col",{attrs:{cols:"24",sm:"12",md:"6"}},[e("v-switch",{attrs:{label:"Autostart"},model:{value:t.autonav,callback:function(a){t.autonav=a},expression:"autonav"}}),e("v-text-field",{attrs:{label:"URL",placeholder:"https://"},model:{value:t.target,callback:function(a){t.target=a},expression:"target"}}),e("v-text-field",{attrs:{label:"Status",placeholder:"slideshow01"},model:{value:t.nextstatus,callback:function(a){t.nextstatus=a},expression:"nextstatus"}})],1),e("v-col",{attrs:{cols:"24",sm:"6",md:"3"}},[e("v-btn",{staticClass:"mx-2",attrs:{fab:"",dark:"",color:"blue"},on:{click:t.sendToDisplay}},[e("v-icon",{attrs:{dark:""}},[t._v(t._s(t.uploadIcon))])],1)],1)],1)],1),e("v-snackbar",{attrs:{timeout:t.snackbartimeout,color:t.snackbarcolor},model:{value:t.snackbar,callback:function(a){t.snackbar=a},expression:"snackbar"}},[t._v(" "+t._s(t.snackbartext)+" "),e("v-btn",{attrs:{text:""},on:{click:function(a){t.snackbarauto=!1}}},[t._v(" Close ")])],1)],1)},A=[],L=(e("caad"),e("2532"),{name:"UploadURL",props:{name:{type:String}},components:{},data:function(){return{target:"",uploadIcon:f["c"],tab:null,nextstatus:"manual01",autonav:!1,snackbartext:"",snackbarcolor:"blue",snackbar:!1,snackbartimeout:2e3}},watch:{},methods:{postAutonav:function(t){var a=this;i.a.post(a.$controllerBase+"/kvstore/"+a.name+"/autonav",{url:t,nextstatus:a.nextstatus}).then((function(t){a.snackbartext="autostart set: "+JSON.stringify(t.data),a.snackbarcolor="blue",a.snackbartimeout=2e3,a.snackbar=!0})).catch((function(t){a.snackbartext="error setting autostart: "+t,a.snackbarcolor="error",a.snackbartimeout=4e3,a.snackbar=!0}))},postNavigate:function(t){var a=this;i.a.post(this.$controllerBase+"/client/"+a.name+"/navigate",{url:t,nextstatus:a.nextstatus}).then((function(t){a.snackbartext="url sent: "+JSON.stringify(t.data),a.snackbarcolor="blue",a.snackbartimeout=2e3,a.snackbar=!0})).catch((function(t){a.snackbartext="error sending url: "+t,a.snackbarcolor="error",a.snackbartimeout=4e3,a.snackbar=!0}))},sendToDisplay:function(){if(this.target.includes("://")){var t=this;this.postNavigate(this.target),t.autonav&&this.postAutonav(this.target)}}}}),U=L,P=(e("928c"),e("4bd4")),D=e("2db4"),E=e("b73d"),F=e("8654"),H=Object(g["a"])(U,J,A,!1,null,"6705489a",null),q=H.exports;w()(H,{VBtn:y["a"],VCol:_["a"],VContainer:V["a"],VForm:P["a"],VIcon:C["a"],VRow:S["a"],VSnackbar:D["a"],VSwitch:E["a"],VTextField:F["a"]});var z=function(){var t=this,a=t.$createElement,e=t._self._c||a;return e("v-form",[e("v-container",[e("v-row",{attrs:{fluid:""}},[e("v-col",{attrs:{cols:"24",sm:"12",md:"6"}},[e("v-switch",{attrs:{label:"Autostart"},model:{value:t.autonav,callback:function(a){t.autonav=a},expression:"autonav"}}),e("v-select",{attrs:{items:t.controllers,"item-text":"InstanceName",label:"Controller",dense:"",solo:""},model:{value:t.controller,callback:function(a){t.controller=a},expression:"controller"}}),e("v-select",{attrs:{items:t.templates,label:"Template",dense:"",solo:""},model:{value:t.template,callback:function(a){t.template=a},expression:"template"}}),e("v-text-field",{attrs:{label:"Status",placeholder:"slideshow01"},model:{value:t.nextstatus,callback:function(a){t.nextstatus=a},expression:"nextstatus"}})],1),e("v-col",{staticStyle:{"vertical-align":"bottom"},attrs:{align:"end",cols:"24",sm:"6",md:"3"}},[e("v-btn",{staticClass:"mx-2",attrs:{fab:"",dark:"",color:"blue"},on:{click:t.sendToDisplay}},[e("v-icon",{attrs:{dark:""}},[t._v(t._s(t.uploadIcon))])],1)],1)],1),e("v-jsoneditor",{attrs:{plus:!1,height:"400px"},model:{value:t.targetJson,callback:function(a){t.targetJson=a},expression:"targetJson"}})],1),e("v-snackbar",{attrs:{timeout:t.snackbartimeout,color:t.snackbarcolor},model:{value:t.snackbar,callback:function(a){t.snackbar=a},expression:"snackbar"}},[t._v(" "+t._s(t.snackbartext)+" "),e("v-btn",{attrs:{text:""},on:{click:function(a){t.snackbar=!1}}},[t._v(" Close ")])],1)],1)},G=[],K=e("27f7"),Q={name:"UploadURL",props:{name:{type:String}},components:{VJsoneditor:K["a"]},data:function(){return{target:"",nextstatus:"manual01",uploadIcon:f["c"],tab:null,controllers:[],controller:"",templates:[],template:"",autonav:!1,snackbartext:"",snackbarcolor:"blue",snackbar:!1,snackbartimeout:2e3}},watch:{name:function(){},controller:function(){var t=this;i.a.get(this.$controllerBase+"/controller/"+this.controller+"/templates").then((function(a){t.templates=a.data}))}},mounted:function(){var t=this;i.a.get(this.$controllerBase+"/controller").then((function(a){t.controllers=a.data}))},methods:{postAutonav:function(t){var a=this;i.a.post(a.$controllerBase+"/kvstore/"+a.name+"/autonav",{url:t,nextstatus:a.nextstatus}).then((function(t){a.snackbartext="autostart set: "+JSON.stringify(t.data),a.snackbarcolor="blue",a.snackbartimeout=2e3,a.snackbar=!0})).catch((function(t){a.snackbartext="error setting autostart: "+t,a.snackbarcolor="error",a.snackbartimeout=4e3,a.snackbar=!0}))},postNavigate:function(t){var a=this;i.a.post(this.$controllerBase+"/client/"+a.name+"/navigate",{url:t,nextstatus:a.nextstatus}).then((function(t){a.snackbartext="url sent: "+JSON.stringify(t.data),a.snackbarcolor="blue",a.snackbartimeout=2e3,a.snackbar=!0})).catch((function(t){a.snackbartext="error sending url: "+t,a.snackbarcolor="error",a.snackbartimeout=4e3,a.snackbar=!0}))},postNavigationFull:function(t){var a=this;i.a.post(a.$controllerBase+"/kvstore/"+a.name+"/"+a.template,a.targetJson).then((function(e){a.snackbartext="template data sent: "+JSON.stringify(e.data),a.snackbarcolor="blue",a.snackbartimeout=2e3,a.snackbar=!0,a.postNavigate(t),a.autonav&&a.postAutonav(t)})).catch((function(t){a.snackbarautotext="error sending template data to kvstore "+a.name+": "+t,a.snackbarurlcolor="error",a.snackbartimeout=4e3,a.snackbarauto=!0}))},sendToDisplay:function(){var t=this;i.a.get(t.$controllerBase+"/client/"+this.name+"/addr").then((function(a){var e=a.data,n="https://"+e+"/"+t.controller+"/templates/"+t.template;t.postNavigationFull(n)})).catch((function(a){t.snackbarautotext="error getting https addr of "+t.name+": "+a,t.snackbarurlcolor="error",t.snackbartimeout=4e3,t.snackbarauto=!0}))}}},W=Q,X=e("b974"),Y=Object(g["a"])(W,z,G,!1,null,"5225427a",null),Z=Y.exports;w()(Y,{VBtn:y["a"],VCol:_["a"],VContainer:V["a"],VForm:P["a"],VIcon:C["a"],VRow:S["a"],VSelect:X["a"],VSnackbar:D["a"],VSwitch:E["a"],VTextField:F["a"]});var tt={name:"Machine",props:{name:{type:String}},components:{UploadURL:q,Screenshot:B,Console:R,Template:Z},data:function(){return{target:"",targetJson:{},tab:null}},watch:{},methods:{}},at=tt,et=(e("d4c7"),e("71a3")),nt=e("c671"),ot=e("fe57"),rt=e("aac8"),st=Object(g["a"])(at,v,b,!1,null,"4ed5ad7c",null),ct=st.exports;w()(st,{VTab:et["a"],VTabItem:nt["a"],VTabs:ot["a"],VTabsItems:rt["a"]});var lt={name:"BRemoteMain",components:{Machine:ct},data:function(){return{drawer:!0,machineIcon:f["a"],clients:[],currMachine:""}},computed:{orderedClients:function(){return d.a.orderBy(this.clients,"InstanceName")}},methods:{selectMachine:function(t){this.currMachine=t}},mounted:function(){var t=this;i.a.get(this.$controllerBase+"/client").then((function(a){return t.clients=a.data}))}},it=lt,ut=e("40dc"),dt=e("5bc1"),vt=e("a75b"),bt=e("ce7e"),mt=e("8860"),pt=e("da13"),ft=e("1800"),ht=e("5d23"),kt=e("f774"),gt=e("71d9"),xt=e("2a7f"),wt=Object(g["a"])(it,s,c,!1,null,null,null),yt=wt.exports;w()(wt,{VAppBar:ut["a"],VAppBarNavIcon:dt["a"],VContent:vt["a"],VDivider:bt["a"],VIcon:C["a"],VList:mt["a"],VListItem:pt["a"],VListItemAction:ft["a"],VListItemContent:ht["a"],VListItemTitle:ht["b"],VNavigationDrawer:kt["a"],VToolbar:gt["a"],VToolbarTitle:xt["a"]});var _t={name:"App",components:{BRemoteMain:yt},data:function(){return{drawer:!0}}},Vt=_t,Ct=e("7496"),St=Object(g["a"])(Vt,o,r,!1,null,null,null),It=St.exports;w()(St,{VApp:Ct["a"]});var Bt=e("f309");n["a"].use(Bt["a"]);var Ot=new Bt["a"]({icons:{iconfont:"mdiSvg"}}),Tt=e("dc60");n["a"].config.productionTip=!1;var $t="";n["a"].use(Tt["a"],{globals:{$controllerBase:$t}}),new n["a"]({vuetify:Ot,render:function(t){return t(It)}}).$mount("#app")},"687c":function(t,a,e){},"81c6":function(t,a,e){"use strict";var n=e("eb22"),o=e.n(n);o.a},"8a3a":function(t,a,e){},"928c":function(t,a,e){"use strict";var n=e("687c"),o=e.n(n);o.a},d4c7:function(t,a,e){"use strict";var n=e("5114"),o=e.n(n);o.a},eb22:function(t,a,e){}});
//# sourceMappingURL=app.3f293b07.js.map