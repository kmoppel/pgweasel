use crate::{aggregators::Aggregator, parsers::LogLine};

pub struct ZeroAgregator;

impl Aggregator for ZeroAgregator {
    fn add(&mut self, log_line: LogLine) {
        println!("{}", log_line.raw);
    }

    fn print(&mut self) {}
}
