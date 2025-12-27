use std::{any::Any, cmp::Reverse, collections::BinaryHeap, time::Duration};

use crate::{aggregators::Aggregator, duration::extract_duration, format::Format};

#[derive(Clone)]
pub struct TopSlowQueries {
    limit: usize,
    heap: BinaryHeap<Reverse<(Duration, Vec<u8>)>>,
}

impl TopSlowQueries {
    pub fn new(limit: usize) -> Self {
        Self {
            limit,
            heap: BinaryHeap::with_capacity(limit),
        }
    }
}

impl Aggregator for TopSlowQueries {
    fn update(&mut self, record: &[u8], _: &Format) {
        let Some(duration) = extract_duration(record) else {
            return;
        };

        if self.heap.len() < self.limit {
            self.heap.push(Reverse((duration, record.to_vec())));
            return;
        }

        if let Some(Reverse((min, _))) = self.heap.peek()
            && duration > *min
        {
            self.heap.pop();
            self.heap.push(Reverse((duration, record.to_vec())));
        }
    }

    fn merge_box(&mut self, other: &dyn Aggregator) {
        let other = other
            .as_any()
            .downcast_ref::<TopSlowQueries>()
            .expect("Aggregator type mismatch");

        for Reverse((duration, record)) in &other.heap {
            if self.heap.len() < self.limit {
                self.heap.push(Reverse((*duration, record.clone())));
            } else if let Some(Reverse((min, _))) = self.heap.peek()
                && *duration > *min
            {
                self.heap.pop();
                self.heap.push(Reverse((*duration, record.clone())));
            }
        }
    }

    fn print(&mut self) {
        let mut items: Vec<_> = self.heap.drain().collect();
        items.sort_by_key(|Reverse((d, _))| *d);

        println!("Top {} slowest queries:", items.len());
        for Reverse((duration, record)) in items.into_iter().rev() {
            println!("--- {:?} ---", duration);
            println!("{}", unsafe { std::str::from_utf8_unchecked(&record) });
        }
    }

    fn boxed_clone(&self) -> Box<dyn Aggregator> {
        Box::new(self.clone())
    }

    fn as_any(&self) -> &dyn Any {
        self
    }
}
