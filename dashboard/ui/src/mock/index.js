import Mock from 'mockjs'

Mock.mock('/api/config/robot_pro', () => {
  return JSON.parse('{"code":0,"error":"","data":[{"name":"/global.conf","sections":[{"name":"DEFAULT","keys":{"cluster":{"type":"string","raw_key":"clusterstring","raw_value":"robot_test","key":"cluster"},"counter_url":{"type":"string","raw_key":"counter_urlstring","raw_value":"http://counter.test.huajiao.com/counter/increase","key":"counter_url"},"debug":{"type":"bool","raw_key":"debugbool","raw_value":"true","key":"debug"},"default_lang":{"type":"string","raw_key":"default_langstring","raw_value":"vi","key":"default_lang"},"focus_url":{"type":"string","raw_key":"focus_urlstring","raw_value":"","key":"focus_url"},"lang_supports":{"type":"[]string","raw_key":"lang_supports[]string","raw_value":"en,vi,zh_cn","key":"lang_supports"},"log_level":{"type":"int","raw_key":"log_levelint","raw_value":"0","key":"log_level"},"pepper_keys":{"type":"map[string]string","raw_key":"pepper_keysmap[string]string","raw_value":"counter.test.huajiao.com:eac63e66d8c4a6f0303f00bc76d0217c","key":"pepper_keys"},"redis_addrs":{"type":"[]string","raw_key":"redis_addrs[]string","raw_value":"10.142.99.152:6511:9a3325ff22e638de2b52b7cd04b15f02","key":"redis_addrs"},"robot_rpcs":{"type":"[]string","raw_key":"robot_rpcs[]string","raw_value":"","key":"robot_rpcs"}}}]},{"name":"/robot.conf","sections":[{"name":"DEFAULT","keys":{"admin_listen":{"type":"string","raw_key":"admin_listenstring","raw_value":":17100","key":"admin_listen"},"gorpc_listen":{"type":"string","raw_key":"gorpc_listenstring","raw_value":":7440","key":"gorpc_listen"},"listens":{"type":"[]string","raw_key":"listens[]string","raw_value":":80","key":"listens"}}}]}]}')
})

Mock.mock('/api/keeper/domains', () => {
  return JSON.parse('{"code":0,"error":"","data":[{"keeper":"127.0.0.1:17000","domain":"robot_test","component":"limb"},{"keeper":"127.0.0.1:17000","domain":"robot_pro","component":"cebera"}]}')
})

Mock.mock('/api/keeper/index', () => {
  return JSON.parse('{"code":0,"error":"","data":["127.0.0.1:17000"]}')
})
