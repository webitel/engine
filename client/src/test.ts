import {Client} from './client'

const cli = new Client({
    endpoint: "http://10.10.10.25:10025",
    token: "MY_TOKEN"
});


cli.connect().then(
 async () => {
     await cli.auth();
     await cli.subscribe("subscribe_self_calls", {ids: [1,2,3,4,56]});
 }
);
