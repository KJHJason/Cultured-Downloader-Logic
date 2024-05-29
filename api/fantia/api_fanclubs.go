package fantia

func (f *FantiaDl) GetFanclubsPosts(dlOptions *FantiaDlOptions) []error {
	return f.GetFanclubsContents(f.FanclubIds, f.FanclubPageNums, POSTS, dlOptions)
}
