INSERT INTO docs (uuid, name, doc) VALUES
    (gen_random_uuid(), 'SOP Refund', 'Prosedur refund: 1. Verifikasi order ID dan status pembatalan. 2. Cek apakah pembayaran sudah diterima. 3. Jika sudah, proses refund dalam 3-5 hari kerja. 4. Jika belum, eskalasi ke tim payment. 5. Notifikasi customer melalui chat.'),
    (gen_random_uuid(), 'SOP Shipping Complaint', 'Prosedur komplain pengiriman: 1. Cek tracking ID di sistem kurir. 2. Jika barang masih dalam perjalanan, informasikan estimasi. 3. Jika barang hilang, buat case untuk tim shipping. 4. Jika barang rusak, minta foto bukti dan buat case.')
ON CONFLICT DO NOTHING;
