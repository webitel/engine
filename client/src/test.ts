import {Client} from './client'
import {Call, CallState} from "./call";

const cli = new Client({
    endpoint: "ws://10.10.10.25:10025",
    token: "USER_TOKEN"
});

cli.connect().then(
 async () => {

     const calls = document.getElementById("calls");

     const newCall = (call : Call) => {
         const el = document.createElement("li")
         el.id = call.id

         const info = document.createElement("div")
         info.id = `${call.id}-info`
         info.textContent = call.toString()

         const btnGroup = document.createElement("div")
         btnGroup.id = `${call.id}-btn`

         const hangupBtn = document.createElement("button")
         hangupBtn.textContent = "hangup"
         hangupBtn.addEventListener("click", () => {
             call.hangup()
         })

         const holdBtn = document.createElement("button")
         holdBtn.textContent = "hold"
         holdBtn.addEventListener("click", () => {
             call.hold()
         })
         const unholdBtn = document.createElement("button")
         unholdBtn.textContent = "un-hold"
         unholdBtn.addEventListener("click", () => {
             call.unHold()
         })

         const dtmfBtn = document.createElement("button")
         dtmfBtn.textContent = "DTMF"
         dtmfBtn.addEventListener("click", () => {
             call.sendDTMF("1")
         })

         const btBtn = document.createElement("button")
         btBtn.textContent = "B Tran"
         btBtn.addEventListener("click", () => {
             call.blindTransfer("400")
         })

         const answerBtn = document.createElement("button")
         answerBtn.textContent = "Answer"
         answerBtn.addEventListener("click", () => {
             call.answer()
         })

         btnGroup.appendChild(hangupBtn)
         btnGroup.appendChild(holdBtn)
         btnGroup.appendChild(unholdBtn)
         btnGroup.appendChild(dtmfBtn)
         btnGroup.appendChild(btBtn)
         btnGroup.appendChild(answerBtn)

         el.appendChild(info)
         el.appendChild(btnGroup)

         // const hangupBtn = document.createElement("button");
         // hangupBtn.textContent = 'Hangup'
         // el.appendChild(hangupBtn)

         calls.appendChild(el)
     };
     const removeCall = (call : Call) => {
         const el = document.getElementById(call.id)
         el.remove()
     };
     const updateCall = (call : Call) => {
         const el = document.getElementById(`${call.id}-info`)
         el.textContent = call.toString();
     };

     const callHandler = (call: Call) => {

         switch (call.state) {
             case CallState.Ringing:
                 newCall(call)
                 console.error("RING RING MF")
                 break;
             case CallState.Active:
                 updateCall(call)
                 console.error("CALL IS ACTIVE")
                 break;
             case CallState.Hold:
                 updateCall(call)
                 console.error("CALL IS HOLD")
                 break;
             case CallState.Hangup:
                 console.error("CALL IS HANGUP", call.hangupCause)
                 updateCall(call)
                 setTimeout(() => {
                     removeCall(call)
                 }, 4000)
                 break;
         }
         // call.hangup();
     }

     await cli.auth();
     await cli.subscribe("call", callHandler, {ids: [1,2,3,4,56]});
     // cli.unSubscribe("call", callHandler)

     // const result = await cli.makeOutboundCall({
     //     toNumber: "123",
     //     toName: "Test Call",
     // });
     // console.error(result)
 }
).catch();
