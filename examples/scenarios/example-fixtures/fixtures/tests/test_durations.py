import unittest

from maintainer_tools.durations import parse_duration


class ParseDurationTests(unittest.TestCase):
    def test_parses_milliseconds_without_changing_other_units(self):
        self.assertEqual(parse_duration("250ms"), 0.25)
        self.assertEqual(parse_duration("2s"), 2.0)
        self.assertEqual(parse_duration("3m"), 180.0)


if __name__ == "__main__":
    unittest.main()
