import os
import shutil
import time
from dataclasses import dataclass
import struct
import pickle
import zlib

HEADER_FORMAT = 'i'  # Integer
CRC_FORMAT = 'I'  # Integer
TIMESTAMP_FORMAT = 'Q'  # C long long

HEADER_SIZE = struct.calcsize(HEADER_FORMAT)
CRC_SIZE = struct.calcsize(CRC_FORMAT)
TIMESTAMP_SIZE = struct.calcsize(TIMESTAMP_FORMAT)


@dataclass
class WALHeader:
    size: int
    type: int


def current_time() -> int:
    return int(time.time() * 1000000)


@dataclass
class WAL:
    filename: str

    def __init__(self, filename: str):
        self.file = None
        self.filename = filename
        self._open_file()

    def _open_file(self):
        try:
            self.file = open(self.filename, 'a+b')
        except Exception as e:
            raise RuntimeError(e)

    def append(self, data: any):
        serialized = pickle.dumps(data)
        now = current_time()
        tim = struct.pack(TIMESTAMP_FORMAT, now)

        data = serialized + tim

        payload = data + struct.pack(CRC_FORMAT, self._crc(data))

        payload_size = len(payload)
        serialized_payload_size = struct.pack(HEADER_FORMAT, payload_size)

        self.file.write(serialized_payload_size)
        self.file.write(payload)

        # Write to disk
        self.file.flush()
        os.fsync(self.file.fileno())

    def _crc(self, payload: bytes) -> int:
        """
        Computes CRC for given payload
        """
        return zlib.crc32(payload) & 0xFFFFFFFF  # Ensure it's an unsigned 32-bit integer

    def recover(self) -> list[any]:
        if not self.file:
            return []

        self.file.seek(0)
        recovered = []
        while True:
            header = self.file.read(HEADER_SIZE)

            if not header:
                # EOF
                break

            try:
                size = struct.unpack(HEADER_FORMAT, header)[0]

                entry_bytes = self.file.read(size)

                if len(entry_bytes) != size:
                    print(
                        "ERROR: Entry bytes doesn't match the expected size in the header. Rest of the file maybe "
                        "corrupted")
                    return recovered

                crc = struct.unpack(CRC_FORMAT, entry_bytes[-CRC_SIZE:])[0]
                serialized_time = entry_bytes[-(CRC_SIZE + TIMESTAMP_SIZE): -CRC_SIZE]
                time_part = struct.unpack(TIMESTAMP_FORMAT, serialized_time)[0]

                serialized_payload = entry_bytes[0: -(CRC_SIZE + TIMESTAMP_SIZE)]

                payload_crc = self._crc(serialized_payload + serialized_time)
                if payload_crc != crc:
                    print("ERROR: CRC doesn't match for an entry. The rest of the file maybe corrupted")
                    return recovered

                payload = pickle.loads(serialized_payload)
                recovered.append((payload, time_part))
            except Exception as e:
                raise e
        return recovered

    def close(self):
        if self.file:
            self.file.close()
            self.file = None

    def remove(self):
        if self.filename:
            if os.path.exists(self.filename):
                os.remove(self.filename)

    def truncate(self, n: int):
        self.close()
        with open(self.filename, 'r+b') as f:
            f.truncate(n)
        self._open_file()
        print(f"Log file '{self.file}' truncated to {n} bytes to simulate crash.")

    def _write(self, data: bytes):
        self.file.write(data)
        self.file.flush()
        os.fsync(self.file.fileno())


def main():
    filename = 'wallog.wal'
    wal = WAL(filename)
    wal.append({'a': 'b'})

    wal.close()

    wal = WAL(filename)
    data = wal.recover()
    for i in data:
        print(i)

    wal.remove()


if __name__ == "__main__":
    main()
