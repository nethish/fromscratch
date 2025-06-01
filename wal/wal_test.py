from wal import WAL
import os


def test_wal(num_entries: int = 0):
    filename = 'log.wal'
    wal = WAL(filename)
    l = []

    for i in range(num_entries):
        payload = {'a': 'b', 'iter': i}
        wal.append(payload)
        l.append(payload)

    wal.close()

    wal = WAL(filename)
    data = wal.recover()

    if len(data) != num_entries:
        print("ERROR: Number of entries don't match")
        wal.remove()
        return

    for i in range(num_entries):
        if data[i][0] != l[i]:
            print("ERROR: Data doesn't match at i")
            wal.remove()
            return

    print("TEST ALL ENTRIES SUCCEEDED")
    wal.remove()


def test_wal_corrupted(num_entries: int = 0):
    filename = 'corrupted.wal'
    wal = WAL(filename)
    l = []

    for i in range(num_entries):
        payload = {'a': 'b', 'iter': i}
        wal.append(payload)
        l.append(payload)

    wal.close()

    wal = WAL(filename)

    original_size = os.path.getsize(filename)
    wal.truncate(original_size - 1)
    data = wal.recover()

    if len(data) != num_entries - 1:
        print("ERROR: Number of entries don't match", len(data))
        wal.remove()
        return

    for i in range(num_entries - 1):
        if data[i][0] != l[i]:
            print("ERROR: Data doesn't match at i", data[i][0], l[i])
            wal.remove()
            return

    print("TEST CORRUPTED SUCCEEDED")
    wal.remove()


def test_wal_crc_corrupted(num_entries: int = 0):
    filename = 'corrupted.wal'
    wal = WAL(filename)
    l = []

    for i in range(num_entries):
        payload = {'a': 'b', 'iter': i}
        wal.append(payload)
        l.append(payload)

    wal.close()

    wal = WAL(filename)

    original_size = os.path.getsize(filename)
    wal.truncate(original_size - 1)
    wal._write(bytes('a', encoding='utf-8'))
    data = wal.recover()

    if len(data) != num_entries - 1:
        print("ERROR: Number of entries don't match", len(data))
        wal.remove()
        return

    for i in range(num_entries - 1):
        if data[i][0] != l[i]:
            print("ERROR: Data doesn't match at i", data[i][0], l[i])
            wal.remove()
            return

    print("TEST CRC CORRUPTED SUCCEEDED")
    wal.remove()


if __name__ == "__main__":
    test_wal()
    test_wal(1)
    test_wal(1000)

    test_wal_corrupted(5)
    test_wal_crc_corrupted(5)
