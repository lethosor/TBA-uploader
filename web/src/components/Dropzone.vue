<template>
    <div
        :class="{dropzone: true, active: active, blocked: blocked}"
        @click="onClick"
        @drop.prevent="onDrop"
        @dragenter.prevent="onDragEnter"
        @dragover.prevent="onDragEnter"
        @dragexit.prevent="onDragLeave"
        @dragleave.prevent="onDragLeave"
    >
        <input
            ref="file"
            type="file"
            :accept="accept"
            @change="onFileChange"
        >
        <p class="title">{{ title }}</p>
        <p>{{ filename || "(no file selected)" }}</p>
        <p>{{ message }}</p>
    </div>
</template>

<style scoped>
.dropzone {
    background: #eee;
    border: 1px dashed #bbb;
    min-width: 300px;
    min-height: 4em;
    margin-bottom: 1em;
    padding: 1em;
    text-align: center;
    cursor: pointer;
    word-break: break-word;
}

.dropzone.active {
    cursor: copy;
    background: #bbb;
}

.dropzone.blocked {
    cursor: no-drop;
    background: #ebb;
}

.dropzone input {
    display: none;
}

.dropzone p {
    margin-bottom: 0;
}

.dropzone p.title {
    font-weight: bold;
}
</style>

<script>
export default {
    name: 'Dropzone',
    props: {
        title: {
            type: String,
            default: 'Upload a file',
        },
        accept: {
            type: String,
            default: '',
        },
    },
    data: function() {
        return {
            filename: '',
            message: '',
            active: false,
            blocked: false,
        };
    },
    methods: {
        _getDataTransferItemFromEvent: function(e) {
            if (e.dataTransfer && e.dataTransfer.items) {
                return Array.from(e.dataTransfer.items).filter(function(item) {
                    return item.kind == 'file' && (!this.accept || this.accept == item.type);
                }.bind(this))[0];
            }
        },
        reset: function() {
            this.filename = '';
            this.message = '';
            this.$refs.file.value = '';
        },
        onClick: function() {
            this.active = true;
            this.$refs.file.click();
            this.active = false;
        },
        onDragEnter: function(e) {
            if (this._getDataTransferItemFromEvent(e)) {
                this.active = true;
            }
            else {
                this.blocked = true;
            }
        },
        onDragLeave: function() {
            this.active = false;
            this.blocked = false;
        },
        onDrop: function(e) {
            this.active = false;
            this.blocked = false;
            var item = this._getDataTransferItemFromEvent(e);
            if (!item) {
                return;
            }
            var file = item.getAsFile();
            this.parseFile(file);
        },
        onFileChange: function(e) {
            if (!e.target.files[0]) {
                this.filename = '';
                this.message = '';
                return;
            }
            this.parseFile(e.target.files[0]);
        },
        parseFile: function(file) {
            this.filename = file.name;
            this.message = 'Loading...';
            var reader = new FileReader();
            reader.readAsText(file);
            reader.onload = function(e) {
                this.message = '';
                this.$emit('upload', {
                    name: this.filename,
                    body: e.target.result,
                });
            }.bind(this);
            reader.onerror = function(e) {
                this.message = 'Read failed';
                this.$emit('error', e);
            }.bind(this);
            this.$refs.file.value = '';
        },
    },
};
</script>
