package model

// defaultRecommendCategories 内置推荐分类（分类表为空时初始化）
// 数组顺序即发现页分类的初始展示顺序
var defaultRecommendCategories = []RecommendCategory{
	{Name: "科技资讯", Description: "科技行业新闻与数码资讯"},
	{Name: "技术博客", Description: "开发者技术文章与工程实践"},
	{Name: "财经商业", Description: "财经快讯与商业深度报道"},
	{Name: "国内新闻", Description: "国内时政与社会新闻"},
	{Name: "AI 人工智能", Description: "AI 前沿动态、论文与产品"},
	{Name: "其他", Description: "其他类型内容"},
}

// defaultRecommendSources 内置推荐源（仅用于初始化/补齐，日常通过后台维护）
// 分类内顺序 = 同一分类段落中的书写顺序
// HotRank > 0 进入发现页"热门订阅"横滑区，数字越小越靠前
var defaultRecommendSources = []RecommendSource{
	// —— 科技资讯 ——
	{Title: "36氪", FeedURL: "https://www.36kr.com/feed", SiteURL: "https://36kr.com", Category: "科技资讯", Description: "提供科技与商业领域的资讯、观点和深度报道", HotRank: 1},
	{Title: "爱范儿", FeedURL: "https://www.ifanr.com/feed", SiteURL: "https://www.ifanr.com", Category: "科技资讯", Description: "让未来触手可及，关注消费电子与互联网", HotRank: 2},
	{Title: "极客公园", FeedURL: "https://www.geekpark.net/rss", SiteURL: "https://www.geekpark.net", Category: "科技资讯", Description: "聚焦科技产品与创新公司的深度报道"},
	{Title: "Solidot", FeedURL: "https://www.solidot.org/index.rss", SiteURL: "https://www.solidot.org", Category: "科技资讯", Description: "奇客资讯，报道科技和互联网相关新闻", HotRank: 10},
	{Title: "InfoQ 中文站", FeedURL: "https://www.infoq.cn/feed", SiteURL: "https://www.infoq.cn", Category: "科技资讯", Description: "面向架构师和技术决策者的技术内容平台", HotRank: 6},
	{Title: "IT之家", FeedURL: "https://www.ithome.com/rss/", SiteURL: "https://www.ithome.com", Category: "科技资讯", Description: "科技新闻与数码产品资讯"},
	{Title: "人人都是产品经理", FeedURL: "https://www.woshipm.com/feed", SiteURL: "https://www.woshipm.com", Category: "科技资讯", Description: "产品经理与互联网运营学习社区"},

	// —— 技术博客 ——
	{Title: "掘金", FeedURL: "https://juejin.cn/rss", SiteURL: "https://juejin.cn", Category: "技术博客", Description: "开发者技术社区，前端/后端/AI 优质文章"},
	{Title: "阮一峰的网络日志", FeedURL: "https://www.ruanyifeng.com/blog/atom.xml", SiteURL: "https://www.ruanyifeng.com/blog/", Category: "技术博客", Description: "科技爱好者周刊与 Web 技术文章", HotRank: 4},
	{Title: "美团技术团队", FeedURL: "https://tech.meituan.com/feed/", SiteURL: "https://tech.meituan.com", Category: "技术博客", Description: "美团一线工程实践与技术总结"},
	{Title: "张鑫旭-鑫空间", FeedURL: "https://www.zhangxinxu.com/wordpress/feed/", SiteURL: "https://www.zhangxinxu.com", Category: "技术博客", Description: "Web 前端开发，CSS/HTML/JavaScript 心得"},
	{Title: "全栈应用开发", FeedURL: "https://www.phodal.com/blog/feeds/rss/", SiteURL: "https://www.phodal.com", Category: "技术博客", Description: "Phodal 的全栈开发与架构实践博客"},
	{Title: "博客园", FeedURL: "https://www.cnblogs.com/rss", SiteURL: "https://www.cnblogs.com", Category: "技术博客", Description: "老牌程序员社区，前端精华多"},

	// —— 财经商业 ——
	{Title: "华尔街见闻", FeedURL: "https://dedicated.wallstreetcn.com/rss.xml", SiteURL: "https://wallstreetcn.com", Category: "财经商业", Description: "全球宏观与市场资讯，7x24 财经快讯", HotRank: 8},
	{Title: "虎嗅网", FeedURL: "https://rsshub.dicomp.net/huxiu/article", SiteURL: "https://www.huxiu.com", Category: "财经商业", Description: "商业资讯与观点平台"},
	{Title: "金十数据", FeedURL: "https://rsshub.dicomp.net/jin10", SiteURL: "https://www.jin10.com", Category: "财经商业", Description: "全球财经快讯与数据日历"},
	{Title: "财联社", FeedURL: "https://rsshub.dicomp.net/cls/depth/1000", SiteURL: "https://www.cls.cn", Category: "财经商业", Description: "7x24 财经快讯与深度报道"},
	{Title: "第一财经 - 新闻", FeedURL: "https://rsshub.dicomp.net/yicai/news", SiteURL: "https://www.yicai.com", Category: "财经商业", Description: "财经新闻与商业报道"},
	{Title: "7*24小时全球财经", FeedURL: "https://rsshub.dicomp.net/10jqka/realtimenews", SiteURL: "https://www.10jqka.com.cn", Category: "财经商业", Description: "同花顺全球财经实时快讯"},
	{Title: "钛媒体", FeedURL: "https://www.tmtpost.com/rss", SiteURL: "https://www.tmtpost.com", Category: "财经商业", Description: "科技与商业深度报道"},
	{Title: "福布斯中文", FeedURL: "https://www.forbeschina.com/rss", SiteURL: "https://www.forbeschina.com", Category: "财经商业", Description: "福布斯中国商业资讯"},
	{Title: "财富中文网", FeedURL: "https://plink.anyfeeder.com/fortunechina", SiteURL: "https://www.fortunechina.com", Category: "财经商业", Description: "财富中文网，全球商业财经报道与榜单"},
	{Title: "雪球", FeedURL: "https://plink.anyfeeder.com/weixin/xueqiujinghua", SiteURL: "https://xueqiu.com", Category: "财经商业", Description: "投资交流社区，股市行情与投资者观点"},

	// —— 国内新闻 ——
	{Title: "央视新闻", FeedURL: "https://plink.anyfeeder.com/weixin/cctvnewscenter", SiteURL: "https://news.cctv.com", Category: "国内新闻", Description: "央视新闻，权威国内外时政要闻"},
	{Title: "头条-求是网", FeedURL: "https://plink.anyfeeder.com/qstheory", SiteURL: "http://www.qstheory.cn", Category: "国内新闻", Description: "求是网头条与理论文章"},
	{Title: "新华社新闻-新华网", FeedURL: "https://plink.anyfeeder.com/newscn/whxw", SiteURL: "http://www.news.cn", Category: "国内新闻", Description: "新华社即时新闻报道"},
	{Title: "澎湃新闻", FeedURL: "https://plink.anyfeeder.com/thepaper", SiteURL: "https://www.thepaper.cn", Category: "国内新闻", Description: "时政与社会深度报道"},
	{Title: "新浪体育", FeedURL: "https://plink.anyfeeder.com/weixin/sports_sina", SiteURL: "https://sports.sina.com.cn", Category: "国内新闻", Description: "新浪体育，国内外体育赛事资讯"},

	// —— AI 人工智能 ——
	{Title: "机器之心", FeedURL: "https://www.jiqizhixin.com/rss", SiteURL: "https://www.jiqizhixin.com", Category: "AI 人工智能", Description: "国内最大 AI 媒体，论文解读 + 产业动态", HotRank: 5},
	{Title: "Rundown AI", FeedURL: "https://www.therundown.ai/feed", SiteURL: "https://www.therundown.ai", Category: "AI 人工智能", Description: "每天一封 AI 要闻简报"},
	{Title: "OpenAI Blog", FeedURL: "https://openai.com/blog/rss.xml", SiteURL: "https://openai.com/blog", Category: "AI 人工智能", Description: "OpenAI 官方博客，模型发布公告与研究文章"},
	{Title: "arXiv cs.AI", FeedURL: "https://rss.arxiv.org/rss/cs.AI", SiteURL: "https://arxiv.org/list/cs.AI/recent", Category: "AI 人工智能", Description: "AI 方向每日新预印本直出"},

	// —— 其他 ——
	{Title: "百度热搜", FeedURL: "https://rss.aishort.top/?type=baidu", SiteURL: "https://top.baidu.com", Category: "其他", Description: "百度搜索实时热点"},
	{Title: "少数派", FeedURL: "https://sspai.com/feed", SiteURL: "https://sspai.com", Category: "其他", Description: "高品质数字消费指南，效率工具与工作流", HotRank: 3},
	{Title: "游研社", FeedURL: "https://www.yystv.cn/rss/feed", SiteURL: "https://www.yystv.cn", Category: "其他", Description: "游戏文化、怀旧考据与深度报道"},
	{Title: "触乐", FeedURL: "https://rsshub.dicomp.net/chuapp/news", SiteURL: "https://www.chuapp.com", Category: "其他", Description: "游戏产业与玩家文化的深度报道"},
	{Title: "机核", FeedURL: "https://www.gcores.com/rss", SiteURL: "https://www.gcores.com", Category: "其他", Description: "游戏、电台与亚文化内容社区"},
}
