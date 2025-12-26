use std::time::Duration;

use crate::{duration::extract_duration, filters::Filter, format::Format};

#[derive(Clone)]
pub struct FilterSlow {
    treshold: Duration,
}

impl FilterSlow {
    pub fn new(treshold: Duration) -> Self {
        FilterSlow { treshold }
    }
}

impl Filter for FilterSlow {
    fn matches(&self, record: &[u8], _fmt: &Format) -> bool {
        if let Some(duration) = extract_duration(record) {
            if duration > self.treshold {
                return true;
            }
        }
        false
    }
}
