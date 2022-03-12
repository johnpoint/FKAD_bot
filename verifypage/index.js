const html = `<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0,maximum-scale=1.0, user-scalable=no">
    <meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1">
    <meta name="renderer" content="webkit">
    <meta http-equiv="Cache-Control" content="no-siteapp">
    <title>人机检验</title>
</head>
<body>
<p>欢迎加入</p>
<code id="verCode">{%code}</code>
<p>以上是您的验证码，请回到群里面直接发送</p>
</body>
</html>`
const setKey = ""


addEventListener('fetch', event => {
    event.respondWith(handleRequest(event.request))
})

/**
 * Respond with hello worker text
 * @param {Request} request
 */
async function handleRequest(request) {
    let data = request.url.split("/")
    if (data.length >= 4) {
        switch (data[3]) {
            case setKey:
                FKAD_KV.put(data[4], data[5], {expirationTtl: 60 * 60})
                break;

            default:
                return new Response(html.replace("{%code}", await FKAD_KV.get(data[3])), {
                    headers: {
                        'content-type': 'text/html;charset=UTF-8',
                    }
                })
        }
    }

    return new Response(JSON.stringify(data), {
        headers: {'content-type': 'text/plain'},
    })
}