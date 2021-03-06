<template>
    <div>
        <div
                class="targets-container"
                ref="container"
                :style="{
                    borderColor: color ? `#${color}`  : '',
                    borderStyle: visible ? 'solid' : 'dashed',
                    opacity: visible ? 1 : 0.7,
                }"
        >
            <div class="picture">
                <div class="picture-container"
                     :style="{ backgroundImage: getBackgroundUrl() }"
                >

                </div>
            </div>
            <div class="name">
                <div class="name-container">
                    <div>{{ name }}</div>
                    <div class="alt">{{ title }}</div>
                </div>
                <div class="you" :class="{ selected: uuid===objectTargetUUID }" v-if="uuid===playerInfo.uuid">you</div>
            </div>
            <div
                class="overlay"
                @mousemove="handleMouseMove"
                @mousedown="handleMouseDown"
                @mouseup="handleMouseUp"
                @mouseleave="handleMouseLeave"
                @dblclick="handleDoubleClick"
                @contextmenu.stop.prevent="onContextMenu"
                @dragenter="handleDragEnter"
                @dragleave="handleDragLeave"
                @drop="handleDrop"
                @dragover.prevent
            ></div>
        </div>
    </div>
</template>

<script>
const OBJECT_TYPE_CHARACTER = 0;
const OBJECT_TYPE_MOB = 1;
const OBJECT_TYPE_ITEM = 2;
import {mapState} from 'vuex';

export default {
    name: 'Target',
    props: ['uuid', 'name', 'objectType', 'pictureKey', 'title', 'color', 'visible'],
    computed: {
        ...mapState([
             'isProduction',
             'objectTargetUUID',
             'playerInfo',
             'itemTooltipMouseCoords',
             'itemTooltipUUID',
             'itemTooltipVisible'
         ]),
    },
    watch: {
        objectTargetUUID: function(target) {
            if (this.uuid === target) {
                this.$refs['container'].classList.add('selected');
            } else {
                this.$refs['container'].classList.remove('selected');
            }
        },
    },
    mounted() {
        switch(this.objectType) {
            case OBJECT_TYPE_CHARACTER:
                this.$refs['container'].classList.add('is-character');
                break;
            case OBJECT_TYPE_MOB:
                this.$refs['container'].classList.add('is-mob');
                break;
            case OBJECT_TYPE_ITEM:
                this.$refs['container'].classList.add('is-item');
                if (this.color.length > 0) {
                    this.$refs['container'].style.borderColor = this.color;
                }
                break;
        }

        if (this.uuid === this.objectTargetUUID) {
            this.$refs['container'].classList.add('selected');
        }
    },
    methods: {
        getBackgroundUrl() {
            if (!this.isProduction) {
                return `url(http://${window.location.hostname}:8081/oi/${this.pictureKey})`;
            }

            return `url(/oi/${this.pictureKey})`;
        },

        handleMouseMove: function(e) {
            if (this.objectType !== OBJECT_TYPE_ITEM) {
                return;
            }

            if (this.itemTooltipMouseCoords.x !== e.clientX || this.itemTooltipMouseCoords.y !== e.clientY) {
                this.$store.dispatch('moveItemTooltip', { x: e.clientX, y: e.clientY });
            }

            if (this.itemTooltipUUID !== this.uuid) {
                this.$store.dispatch('showItemTooltip', this.uuid);
            }
        },

        handleMouseLeave: function() {
            this.hideTooltip();
        },

        handleMouseDown: function() {
            this.$refs['container'].classList.add('mouse-down');
        },

        handleMouseUp: function(e) {
            this.$refs['container'].classList.remove('mouse-down');
            if (this.$store.state.permissions.indexOf('CAN_BUILD') >= 0) {
                if (e.shiftKey) {
                    if (this.objectType === OBJECT_TYPE_CHARACTER) {
                        this.$store.dispatch('sendSlashCommand', {
                            command: `/character edit "${this.name}"`,
                            hidden: true,
                        });
                    } else if (this.objectType === OBJECT_TYPE_MOB) {
                        this.$store.dispatch('sendSlashCommand', {
                            command: `/mob iedit "${this.uuid}"`,
                            hidden: true,
                        });
                    } else if (this.objectType === OBJECT_TYPE_ITEM) {
                        this.$store.dispatch('sendSlashCommand', {
                            command: `/item iedit "${this.uuid}"`,
                            hidden: true,
                        });
                    }
                    return
                }
            }

            this.$store.dispatch('setObjectTarget', this.uuid);
        },

        hideTooltip: function() {
            if (this.itemTooltipVisible) {
                this.$store.dispatch('hideItemTooltip');
            }
        },

        handleDoubleClick: function() {
            switch(this.objectType) {
                case OBJECT_TYPE_ITEM:
                    this.$store.dispatch('sendSlashCommand', {
                        command: `/get "${this.uuid}"`,
                        hidden: true,
                    });
                    this.$store.dispatch('setObjectTarget', '');
                    this.hideTooltip();
                    break;
                case OBJECT_TYPE_MOB:
                    this.$store.dispatch('sendSlashCommand', {
                        command: `/interact "${this.uuid}"`,
                        hidden: true,
                    });
                    this.$store.dispatch('setObjectTarget', '');
                    break;
            }
        },

        onContextMenu: function(e) {
            const menu = [];

            switch(this.objectType) {
                case OBJECT_TYPE_CHARACTER:
                    menu.push(`Look %s|/look ${this.uuid}`);
                    menu.push(`Edit %s|/character edit ${this.name}||CAN_CHAREDIT`);
                    break;
                case OBJECT_TYPE_ITEM:
                    menu.push(`Look %s|/look ${this.uuid}`);
                    menu.push(`Wiki %s|wiki:/items/%s`);
                    menu.push(`Pickup %s|/get ${this.uuid}`);
                    menu.push(`Edit %s|/item iedit ${this.uuid}||CAN_BUILD`);
                    menu.push(`Edit-Parent %s|/item edit ${this.name}||CAN_BUILD`);
                    menu.push(`Destroy %s|/destroy ${this.uuid}||CAN_BUILD`);
                    break;
                case OBJECT_TYPE_MOB:
                    menu.push(`Look %s|/look ${this.uuid}`);
                    menu.push(`Interact %s|/interact ${this.uuid}`);
                    menu.push(`Edit %s|/mob iedit ${this.uuid}||CAN_BUILD`);
                    menu.push(`Edit-Parent %s|/mob edit ${this.name}||CAN_BUILD`);
                    menu.push(`Destroy %s|/destroy ${this.uuid}||CAN_BUILD`);
                    break;
            }

            this.$store.dispatch(
                'showContextMenu',
                {
                    object: {
                        name: this.name,
                        color: `#${this.color}`,
                        subjectBrackets: (this.objectType === OBJECT_TYPE_ITEM),
                    },
                    at: {
                        x: e.pageX,
                        y: e.pageY,
                    },
                    items: menu,
                }
            );
        },

        handleDragEnter: function() {
            this.$refs['container'].classList.add('can-drop-item');
        },

        handleDragLeave: function() {
            this.$refs['container'].classList.remove('can-drop-item');
        },

        handleDrop: function(e) {
            this.$refs['container'].classList.remove('can-drop-item');
            const item_uuid = e.dataTransfer.getData('item_uuid');
            if (item_uuid) {
                this.$store.dispatch('sendSlashCommand', {
                    command: `/give "${this.uuid}" "${item_uuid}"`
                });
            }
        }
    }
}
</script>

<style lang="scss" scoped>
.targets-container {
    transition: all .1s ease-in-out;
    transform: scale(1);
    display: flex;
    border-width: 0 !important;

    &.can-drop-item {
         transform: scale(1.1) !important;
    }

    &.selected {
         border: 1px solid #ffeb3b !important;
         background-color: #231f00;
    }

    &.mouse-down {
        transform: scale(1.01) !important;
    }

    &.is-character {
         border: 1px solid #353535;
    }

    &.is-mob {
        border: 1px solid #673604;

        .name {
            color: #d48a3e;
        }
    }

    &.is-item {
         border: 1px solid #fff;

        .name {
            color: #fff;
        }
    }

    &:hover {
        cursor: pointer;
        
    }

    .picture {
        flex-basis: 50px;

        .picture-container {
            height: 50px;
            box-shadow: inset 0px 0px 5px 0px #3a3a3a;
            background-size: contain;
        }
    }

    .name {
        flex-grow: 1;
        display: flex;
        align-items: center;
        margin-left: 10px;

        .name-container {
            font-weight: 600;

            .alt {
                font-weight: 400;
                font-size: 12px;
            }
        }

        .you {
            position: absolute;
            right: -1px;
            top: -1px;
            background-color: #353535;
            padding: 2px 5px;
            border: 1px solid #353535;
            text-transform: uppercase;
            font-size: 12px;
            transition: all .1s ease-in-out;
            border-bottom-left-radius: 3px;

            &.selected {
                 background-color: #eedb38;
                 border: 1px solid #eedb38;
                 color: #000;
            }
        }
    }

    .overlay {
        position: absolute;
        top: 0px;
        left: 0px;
        height: 100%;
        width: 100%;
        z-index: 999;
    }
}
</style>