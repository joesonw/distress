package app

//func parseOutputAsTallyScope(out string, cachedReporters ...tally.CachedStatsReporter) (tally.Scope, io.Closer, error) {
//	u, err := url.Parse(out)
//	if err != nil {
//		return nil, nil, err
//	}
//	q := u.Query()
//
//	var reporter tally.StatsReporter
//	var cachedReporter tally.CachedStatsReporter
//	var pusher *push.Pusher
//
//	switch u.Scheme {
//	case "prometheus+https", "prometheus+http":
//		{
//			user := ""
//			if u.User != nil {
//				user = u.User.String() + "@"
//			}
//			pusher = push.New(strings.Split(u.Scheme, "+")[1]+"://"+user+u.Host+u.Path, "tally")
//			reg := prometheus.NewRegistry()
//			pusher.Gatherer(reg)
//			cachedReporter = tally_prometheus.NewReporter(tally_prometheus.Options{
//				Registerer: reg,
//			})
//		}
//	default:
//		return tally.NoopScope, ioutil.NopCloser(nil), nil
//	}
//
//	freq := time.Second
//	tags := map[string]string{}
//	separator := q.Get("separator")
//	prefix := q.Get("prefix")
//
//	if f := q.Get("every"); f != "" {
//		freq, err = time.ParseDuration(f)
//		if err != nil {
//			return nil, nil, err
//		}
//	}
//
//	for k := range q {
//		if strings.HasPrefix(k, "tag.") {
//			tags[k[4:]] = q.Get(k)
//		}
//	}
//
//	if cachedReporter != nil {
//		cachedReporters = append(cachedReporters, cachedReporter)
//	}
//
//	scope, closer := tally.NewRootScope(tally.ScopeOptions{
//		Tags:           tags,
//		Prefix:         prefix,
//		Reporter:       reporter,
//		CachedReporter: tally_multi.NewMultiCachedReporter(cachedReporters...),
//		Separator:      separator,
//	}, freq)
//	if pusher != nil {
//		closer = prometheusPusherCloser{pusher, closer}
//	}
//	return scope, closer, nil
//}
//
//type prometheusPusherCloser struct {
//	pusher *push.Pusher
//	closer io.Closer
//}
//
//func (c prometheusPusherCloser) Close() error {
//	if err := c.pusher.Push(); err != nil {
//		return err
//	}
//	return c.closer.Close()
//}
