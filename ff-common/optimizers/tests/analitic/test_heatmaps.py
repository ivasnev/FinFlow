"""Проверка координат каждой экспериментальной тепловой карты notebook."""

import json
from pathlib import Path
import unittest

import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt
from matplotlib.backend_bases import MouseEvent
from matplotlib.colors import LogNorm
import numpy as np
import pandas as pd


class HeatmapCoordinatesTests(unittest.TestCase):
    """Проверяет обе ячейки графиков на данных с однозначно различимыми столбцами."""

    def test_all_measured_heatmaps_use_actual_grid(self):
        """Подписи 100 и 1000 попадают в свои пиксели во всех основных и дополнительных картах."""
        notebook = json.loads(Path(__file__).with_name("benchmark_analysis.ipynb").read_text())
        nodes = [5, 10, 25, 50, 100, 200]
        transactions = list(range(10, 101, 10)) + list(range(1000, 10001, 1000))
        algorithms = ["Greedy", "Dinic", "EdmondsKarp", "PushRelabel", "Interactive Dinic",
                      "Interactive EdmondsKarp", "Interactive PushRelabel"]
        rows = []
        for algorithm in algorithms:
            for node in nodes:
                for transaction in transactions:
                    rows.append(dict(algorithm=algorithm, nodes=node, transactions=transaction,
                        ns_per_op=transaction, reduction_len_pct=transaction / 100,
                        allocs_per_op=transaction, bytes_per_op=transaction * 1024,
                        collapse_pct=transaction / 100, input_after_collapse=transaction,
                        collapse_ns_per_op=transaction, optimize_ns_per_op=transaction))
        namespace = dict(plt=plt, np=np, pd=pd, LogNorm=LogNorm, df_all=pd.DataFrame(rows),
                         NODES_SWEEP=nodes, TRANS_SWEEP=transactions)
        checked = []

        def inspect_figure():
            """Сопоставляет подписи и цвета уже построенной фигуры через экранные координаты."""
            figure = plt.gcf()
            figure.canvas.draw()
            for axis in figure.axes:
                for image in axis.images:
                    self.assertEqual(axis.get_xscale(), "linear")
                    self.assertEqual(list(axis.get_xticks()), list(range(len(transactions))))
                    self.assertEqual([tick.get_text() for tick in axis.get_xticklabels()], list(map(str, transactions)))
                    self.assertEqual([tick.get_text() for tick in axis.get_yticklabels()], list(map(str, nodes)))
                    self.assertEqual(list(image.get_extent()), [-0.5, 19.5, 5.5, -0.5])
                    for column in (9, 10):
                        x, y = axis.transData.transform((axis.get_xticks()[column], 2))
                        event = MouseEvent("motion_notify_event", figure.canvas, x, y)
                        self.assertEqual(image.get_cursor_data(event), image.get_array()[2, column])
                    checked.append(axis.get_title())
            plt.close(figure)

        original_show = plt.show
        plt.show = inspect_figure
        try:
            for index in (10, 12):
                exec(compile("".join(notebook["cells"][index]["source"]), f"notebook-cell-{index}", "exec"), namespace)
        finally:
            plt.show = original_show
            plt.close("all")
        self.assertEqual(len(checked), 58, "some measured heatmaps were not checked")


if __name__ == "__main__":
    unittest.main()
