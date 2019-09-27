import {Client} from './client'

const cli = new Client({
    endpoint: "ws://192.168.177.199/ws",
    token: "USER_TOKEN"
});


cli.connect().then(
 async () => {
     await cli.auth();
     await cli.subscribe("subscribe_self_calls", {ids: [1,2,3,4,56]});
 }
);
