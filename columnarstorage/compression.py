class RLECompressor:
    @staticmethod
    def compress(data):
        if not data:
            return []

        compressed = []
        prev = data[0]
        count = 1
        for val in data[1:]:
            if val == prev:
                count += 1
            else:
                compressed.append((prev, count))
                prev = val
                count = 1
        compressed.append((prev, count))
        return compressed

    @staticmethod
    def decompress(compressed):
        result = []
        for value, count in compressed:
            result.extend([value] * count)
        return result


class DictionaryCompressor:
    @staticmethod
    def compress(data):
        dictionary = {}
        reverse_dict = []
        compressed = []
        for val in data:
            if val not in dictionary:
                dictionary[val] = len(reverse_dict)
                reverse_dict.append(val)
            compressed.append(dictionary[val])
        return compressed, reverse_dict

    @staticmethod
    def decompress(compressed, reverse_dict):
        return [reverse_dict[idx] for idx in compressed]
